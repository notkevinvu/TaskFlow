package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	// Create backup filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("backup_%s.sql", timestamp)

	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create backup file: %v", err)
	}
	defer file.Close()

	// Write header
	fmt.Fprintf(file, "-- TaskFlow Database Backup\n")
	fmt.Fprintf(file, "-- Generated: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "-- Database: Supabase PostgreSQL\n\n")

	// Tables to backup (in order due to foreign keys)
	tables := []string{"users", "tasks", "task_history"}

	for _, table := range tables {
		if err := backupTable(ctx, conn, file, table); err != nil {
			log.Printf("Warning: Failed to backup table %s: %v", table, err)
		}
	}

	fmt.Printf("✅ Backup created: %s\n", filename)
}

func backupTable(ctx context.Context, conn *pgx.Conn, file *os.File, table string) error {
	fmt.Fprintf(file, "\n-- Table: %s\n", table)

	// Get column names
	rows, err := conn.Query(ctx, fmt.Sprintf(`
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_name = '%s'
		ORDER BY ordinal_position`, table))
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	var columns []string
	var types []string
	for rows.Next() {
		var col, typ string
		if err := rows.Scan(&col, &typ); err != nil {
			rows.Close()
			return err
		}
		columns = append(columns, col)
		types = append(types, typ)
	}
	rows.Close()

	if len(columns) == 0 {
		fmt.Fprintf(file, "-- Table %s not found or empty schema\n", table)
		return nil
	}

	// Get data
	dataRows, err := conn.Query(ctx, fmt.Sprintf("SELECT * FROM %s", table))
	if err != nil {
		return fmt.Errorf("failed to query table: %w", err)
	}
	defer dataRows.Close()

	rowCount := 0
	for dataRows.Next() {
		values, err := dataRows.Values()
		if err != nil {
			return fmt.Errorf("failed to get row values: %w", err)
		}

		valueStrings := make([]string, len(values))
		for i, v := range values {
			valueStrings[i] = formatValue(v)
		}

		fmt.Fprintf(file, "INSERT INTO %s (%s) VALUES (%s);\n",
			table,
			strings.Join(columns, ", "),
			strings.Join(valueStrings, ", "))
		rowCount++
	}

	fmt.Fprintf(file, "-- %d rows exported from %s\n", rowCount, table)
	fmt.Printf("  📦 %s: %d rows\n", table, rowCount)
	return nil
}

func formatValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}

	switch val := v.(type) {
	case string:
		// Escape single quotes
		escaped := strings.ReplaceAll(val, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case []byte:
		escaped := strings.ReplaceAll(string(val), "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case time.Time:
		return fmt.Sprintf("'%s'", val.Format(time.RFC3339))
	case bool:
		if val {
			return "TRUE"
		}
		return "FALSE"
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", val)
	case []string:
		if len(val) == 0 {
			return "ARRAY[]::text[]"
		}
		escaped := make([]string, len(val))
		for i, s := range val {
			escaped[i] = fmt.Sprintf("'%s'", strings.ReplaceAll(s, "'", "''"))
		}
		return fmt.Sprintf("ARRAY[%s]", strings.Join(escaped, ", "))
	default:
		// For arrays and other types, try string conversion
		str := fmt.Sprintf("%v", val)
		if strings.HasPrefix(str, "[") && strings.HasSuffix(str, "]") {
			// Likely an array, handle it
			return fmt.Sprintf("'%s'", strings.ReplaceAll(str, "'", "''"))
		}
		escaped := strings.ReplaceAll(str, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	}
}
