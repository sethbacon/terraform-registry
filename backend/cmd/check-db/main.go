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

	// Check modules
	fmt.Println("=== MODULES ===")
	rows, err := db.Query("SELECT id, namespace, name, system FROM modules")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, namespace, name, system string
		rows.Scan(&id, &namespace, &name, &system)
		fmt.Printf("Module: %s/%s/%s (ID: %s)\n", namespace, name, system, id)
	}

	// Check versions
	fmt.Println("\n=== MODULE VERSIONS ===")
	rows2, err := db.Query("SELECT id, module_id, version, readme FROM module_versions")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows2.Close()

	count := 0
	for rows2.Next() {
		var id, moduleId, version string
		var readme *string
		rows2.Scan(&id, &moduleId, &version, &readme)
		hasReadme := "NO"
		if readme != nil && *readme != "" {
			hasReadme = fmt.Sprintf("YES (%d chars)", len(*readme))
		}
		fmt.Printf("Version: %s (Module ID: %s, Version ID: %s) - README: %s\n", version, moduleId, id, hasReadme)
		count++
	}

	if count == 0 {
		fmt.Println("No versions found!")
	}
}
