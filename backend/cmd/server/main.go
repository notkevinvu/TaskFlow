package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/notkevinvu/taskflow/backend/internal/config"
	"github.com/notkevinvu/taskflow/backend/internal/handler"
	"github.com/notkevinvu/taskflow/backend/internal/logger"
	"github.com/notkevinvu/taskflow/backend/internal/metrics"
	"github.com/notkevinvu/taskflow/backend/internal/middleware"
	"github.com/notkevinvu/taskflow/backend/internal/ratelimit"
	"github.com/notkevinvu/taskflow/backend/internal/repository"
	"github.com/notkevinvu/taskflow/backend/internal/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize structured logger
	appLogger := logger.New(logger.Config{
		Level:  logger.LogLevel(cfg.LogLevel),
		Format: cfg.LogFormat,
	})
	slog.SetDefault(appLogger)

	slog.Info("Application starting",
		"port", cfg.Port,
		"gin_mode", cfg.GinMode,
		"log_level", cfg.LogLevel,
		"log_format", cfg.LogFormat,
	)

	// Initialize database connection pool
	dbPool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// Verify database connection
	if err := dbPool.Ping(context.Background()); err != nil {
		slog.Error("Failed to ping database", "error", err)
		os.Exit(1)
	}
	slog.Info("Successfully connected to database")

	// Initialize Redis rate limiter (optional - falls back to in-memory if unavailable)
	redisLimiter, err := ratelimit.NewRedisLimiter(cfg.RedisURL)
	if err != nil {
		slog.Warn("Unable to connect to Redis, using in-memory rate limiting", "error", err)
		redisLimiter = nil
	} else {
		defer redisLimiter.Close()
		slog.Info("Successfully connected to Redis for rate limiting")
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(dbPool)
	taskRepo := repository.NewTaskRepository(dbPool)
	taskHistoryRepo := repository.NewTaskHistoryRepository(dbPool)
	taskSeriesRepo := repository.NewTaskSeriesRepository(dbPool)
	userPrefsRepo := repository.NewUserPreferencesRepository(dbPool)
	dependencyRepo := repository.NewDependencyRepository(dbPool)
	templateRepo := repository.NewTaskTemplateRepository(dbPool)
	gamificationRepo := repository.NewGamificationRepository(dbPool)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiryHours)
	taskService := service.NewTaskService(taskRepo, taskHistoryRepo)
	insightsService := service.NewInsightsService(taskRepo)
	recurrenceService := service.NewRecurrenceService(taskRepo, taskSeriesRepo, userPrefsRepo, taskHistoryRepo)
	subtaskService := service.NewSubtaskService(taskRepo, taskHistoryRepo)
	dependencyService := service.NewDependencyService(dependencyRepo, taskRepo)
	templateService := service.NewTaskTemplateService(templateRepo)
	gamificationService := service.NewGamificationService(gamificationRepo, taskRepo)

	// Wire recurrence service into task service for recurring task completion support
	taskService.SetRecurrenceService(recurrenceService)

	// Wire subtask service into task service for parent completion validation
	taskService.SetSubtaskService(subtaskService)

	// Wire dependency service into task service for blocker validation
	taskService.SetDependencyService(dependencyService)

	// Wire gamification service into task service for completion rewards
	taskService.SetGamificationService(gamificationService)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	taskHandler := handler.NewTaskHandler(taskService)
	categoryHandler := handler.NewCategoryHandler(taskService)
	analyticsHandler := handler.NewAnalyticsHandler(taskRepo)
	insightsHandler := handler.NewInsightsHandler(insightsService, taskService)
	recurrenceHandler := handler.NewRecurrenceHandler(recurrenceService)
	subtaskHandler := handler.NewSubtaskHandler(subtaskService)
	dependencyHandler := handler.NewDependencyHandler(dependencyService)
	templateHandler := handler.NewTaskTemplateHandler(templateService)
	gamificationHandler := handler.NewGamificationHandler(gamificationService)

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Initialize router
	router := gin.New() // Use gin.New() instead of Default() to have full control over middleware

	// Apply middleware in order (RequestLogger before ErrorHandler to capture error context)
	router.Use(gin.Recovery())                        // Recover from panics
	router.Use(metrics.Middleware())                  // Prometheus metrics (before other middleware to capture all requests)
	router.Use(middleware.RequestLogger())            // Log all requests with error context
	router.Use(middleware.CORS(cfg.AllowedOrigins))   // CORS
	router.Use(middleware.RateLimiter(redisLimiter, cfg.RateLimitRPM)) // Rate limiting with Redis backend
	router.Use(middleware.ErrorHandler())             // Error handler must be last to catch errors from routes

	// Prometheus metrics endpoint (no auth required for scraping)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		health := gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"services": gin.H{
				"database": "healthy",
			},
		}

		// Check Redis health if available
		if redisLimiter != nil {
			if err := redisLimiter.Health(c.Request.Context()); err != nil {
				health["services"].(gin.H)["redis"] = "unhealthy"
				slog.Warn("Redis health check failed", "error", err)
			} else {
				health["services"].(gin.H)["redis"] = "healthy"
			}
		} else {
			health["services"].(gin.H)["redis"] = "not configured"
		}

		c.JSON(http.StatusOK, health)
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", middleware.AuthRequired(cfg.JWTSecret), authHandler.Me)
		}

		// Task routes (protected)
		tasks := v1.Group("/tasks")
		tasks.Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			tasks.POST("", taskHandler.Create)
			tasks.GET("", taskHandler.List)
			tasks.GET("/calendar", taskHandler.GetCalendar)
			tasks.POST("/suggest-category", insightsHandler.SuggestCategory)
			tasks.POST("/bulk-delete", taskHandler.BulkDelete)
			tasks.POST("/bulk-restore", taskHandler.BulkRestore)
			tasks.GET("/:id", taskHandler.Get)
			tasks.PUT("/:id", taskHandler.Update)
			tasks.DELETE("/:id", taskHandler.Delete)
			tasks.POST("/:id/bump", taskHandler.Bump)
			tasks.POST("/:id/complete", taskHandler.Complete)
			tasks.GET("/:id/estimate", insightsHandler.GetTimeEstimate)
			// Subtask routes (nested under tasks)
			tasks.POST("/:id/subtasks", subtaskHandler.CreateSubtask)
			tasks.GET("/:id/subtasks", subtaskHandler.GetSubtasks)
			tasks.GET("/:id/subtask-info", subtaskHandler.GetSubtaskInfo)
			tasks.GET("/:id/expanded", subtaskHandler.GetTaskExpanded)
			tasks.GET("/:id/can-complete", subtaskHandler.CanCompleteParent)
			// Dependency routes (nested under tasks)
			tasks.POST("/:id/dependencies", dependencyHandler.AddDependency)
			tasks.GET("/:id/dependencies", dependencyHandler.GetDependencyInfo)
			tasks.DELETE("/:id/dependencies/:blocker_id", dependencyHandler.RemoveDependency)
			tasks.GET("/:id/can-complete-dependencies", dependencyHandler.CheckCanComplete)
		}

		// Subtask routes (protected) - for subtask-specific operations
		subtasks := v1.Group("/subtasks")
		subtasks.Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			subtasks.POST("/:id/complete", subtaskHandler.CompleteSubtask)
		}

		// Category routes (protected)
		categories := v1.Group("/categories")
		categories.Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			categories.PUT("/rename", categoryHandler.Rename)
			categories.DELETE("/:name", categoryHandler.Delete)
		}

		// Analytics routes (protected)
		analytics := v1.Group("/analytics")
		analytics.Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			analytics.GET("/summary", analyticsHandler.GetSummary)
			analytics.GET("/trends", analyticsHandler.GetTrends)
			analytics.GET("/heatmap", analyticsHandler.GetProductivityHeatmap)
			analytics.GET("/category-trends", analyticsHandler.GetCategoryTrends)
		}

		// Insights routes (protected)
		insights := v1.Group("/insights")
		insights.Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			insights.GET("", insightsHandler.GetInsights)
		}

		// Series routes (protected) - recurring task series management
		series := v1.Group("/series")
		series.Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			series.GET("", recurrenceHandler.ListSeries)
			series.GET("/:id/history", recurrenceHandler.GetSeriesHistory)
			series.PUT("/:id", recurrenceHandler.UpdateSeries)
			series.POST("/:id/deactivate", recurrenceHandler.DeactivateSeries)
		}

		// Recurrence preferences routes (protected)
		preferences := v1.Group("/preferences/recurrence")
		preferences.Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			preferences.GET("", recurrenceHandler.GetPreferences)
			preferences.GET("/effective", recurrenceHandler.GetEffectiveCalculation)
			preferences.PUT("/default", recurrenceHandler.SetDefaultPreference)
			preferences.PUT("/category/:category", recurrenceHandler.SetCategoryPreference)
			preferences.DELETE("/category/:category", recurrenceHandler.DeleteCategoryPreference)
		}

		// Template routes (protected)
		templates := v1.Group("/templates")
		templates.Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			templates.POST("", templateHandler.CreateTemplate)
			templates.GET("", templateHandler.ListTemplates)
			templates.GET("/:id", templateHandler.GetTemplate)
			templates.PUT("/:id", templateHandler.UpdateTemplate)
			templates.DELETE("/:id", templateHandler.DeleteTemplate)
			templates.POST("/:id/use", templateHandler.UseTemplate)
		}

		// Gamification routes (protected)
		gamification := v1.Group("/gamification")
		gamification.Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			gamification.GET("/dashboard", gamificationHandler.GetDashboard)
			gamification.GET("/stats", gamificationHandler.GetStats)
			gamification.PUT("/timezone", gamificationHandler.SetTimezone)
			gamification.GET("/timezone", gamificationHandler.GetTimezone)
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		slog.Info("Starting HTTP server", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Received shutdown signal, gracefully shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exited successfully")
}
