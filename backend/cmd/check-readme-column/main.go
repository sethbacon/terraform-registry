package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dbPassword := os.Getenv("DATABASE_PASSWORD")
	if dbPassword == "" {
		dbPassword = "registry"
	}

	connStr := fmt.Sprintf("host=localhost port=5432 user=registry password=%s dbname=terraform_registry sslmode=disable", dbPassword)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	// Check if readme column exists
	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name='module_versions' AND column_name='readme'
		)
	`).Scan(&exists)

	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	if exists {
		fmt.Println("✓ README column exists in module_versions table")
	} else {
		fmt.Println("✗ README column NOT found in module_versions table")
		fmt.Println("Running migration...")

		_, err = db.Exec("ALTER TABLE module_versions ADD COLUMN readme TEXT")
		if err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("✓ README column added successfully")
	}
}
