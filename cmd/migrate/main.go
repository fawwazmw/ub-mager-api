package main

import (
	"fmt"
	"log"
	"os"

	"github.com/wardayadev/ub-mager-api/internal/config"
	"github.com/wardayadev/ub-mager-api/internal/database"
	"github.com/wardayadev/ub-mager-api/internal/model"
	"gorm.io/gorm"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./cmd/migrate [up|fresh]")
		fmt.Println("")
		fmt.Println("Commands:")
		fmt.Println("  up      Run auto-migration (create/alter tables)")
		fmt.Println("  fresh   Drop all tables and re-migrate from scratch")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := database.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	switch os.Args[1] {
	case "up":
		migrateUp(db)
	case "fresh":
		migrateFresh(db)
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		fmt.Println("Available: up, fresh")
		os.Exit(1)
	}
}

func migrateUp(db *gorm.DB) {
	fmt.Println("⬆️  Running migrations...")

	if err := db.AutoMigrate(
		&model.User{},
		&model.DriverProfile{},
		&model.Ride{},
		&model.Rating{},
	); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	fmt.Println("✅ Migrations complete")
}

func migrateFresh(db *gorm.DB) {
	fmt.Println("🗑️  Dropping all tables...")

	// Drop in reverse dependency order
	tables := []string{
		"ratings",
		"rides",
		"driver_profiles",
		"users",
	}

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)).Error; err != nil {
			log.Printf("  ⚠️  Failed to drop %s: %v", table, err)
		} else {
			fmt.Printf("  🗑️  Dropped: %s\n", table)
		}
	}

	fmt.Println("")
	migrateUp(db)
}
