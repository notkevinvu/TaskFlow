package main

import (
	"context"
	"fmt"
	"log"
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
	"github.com/notkevinvu/taskflow/backend/internal/middleware"
	"github.com/notkevinvu/taskflow/backend/internal/ratelimit"
	"github.com/notkevinvu/taskflow/backend/internal/repository"
	"github.com/notkevinvu/taskflow/backend/internal/service"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database connection pool
	dbPool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbPool.Close()

	// Verify database connection
	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}
	log.Println("Successfully connected to database")

	// Initialize Redis rate limiter (optional - falls back to allowing all requests if unavailable)
	redisLimiter, err := ratelimit.NewRedisLimiter(cfg.RedisURL)
	if err != nil {
		log.Printf("Warning: Unable to connect to Redis: %v (rate limiting disabled)\n", err)
		redisLimiter = nil
	} else {
		defer redisLimiter.Close()
		log.Println("Successfully connected to Redis for rate limiting")
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(dbPool)
	taskRepo := repository.NewTaskRepository(dbPool)
	taskHistoryRepo := repository.NewTaskHistoryRepository(dbPool)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiryHours)
	taskService := service.NewTaskService(taskRepo, taskHistoryRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	taskHandler := handler.NewTaskHandler(taskService)
	categoryHandler := handler.NewCategoryHandler(taskService)
	analyticsHandler := handler.NewAnalyticsHandler(taskRepo)

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Initialize router
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.CORS(cfg.AllowedOrigins))
	router.Use(middleware.RateLimiter(redisLimiter, cfg.RateLimitRPM))
	router.Use(middleware.ErrorHandler()) // Error handler must be last to catch errors from routes

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"timestamp": time.Now().UTC(),
		})
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
			tasks.GET("/:id", taskHandler.Get)
			tasks.PUT("/:id", taskHandler.Update)
			tasks.DELETE("/:id", taskHandler.Delete)
			tasks.POST("/:id/bump", taskHandler.Bump)
			tasks.POST("/:id/complete", taskHandler.Complete)
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
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on port %s\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v\n", err)
	}

	log.Println("Server exited successfully")
}
