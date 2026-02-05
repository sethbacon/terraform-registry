package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Get database password from environment
	dbPassword := os.Getenv("DATABASE_PASSWORD")
	if dbPassword == "" {
		dbPassword = "registry"
	}

	// Connect to database
	connStr := fmt.Sprintf("host=localhost port=5432 user=registry password=%s dbname=terraform_registry sslmode=disable", dbPassword)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Connected to database successfully")

	// Execute delete statements
	statements := []string{
		"DELETE FROM module_versions",
		"DELETE FROM modules",
		"DELETE FROM provider_platforms",
		"DELETE FROM provider_versions",
		"DELETE FROM providers",
	}

	for _, stmt := range statements {
		result, err := db.Exec(stmt)
		if err != nil {
			log.Printf("Warning executing '%s': %v", stmt, err)
			continue
		}
		rowsAffected, _ := result.RowsAffected()
		fmt.Printf("✓ %s - %d rows deleted\n", stmt, rowsAffected)
	}

	fmt.Println("\n✓ Database cleaned successfully. You can now re-upload modules with README support.")
}
