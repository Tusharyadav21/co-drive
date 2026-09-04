package main

import (
	"flag"
	"log"
	"os"

	"co-drive/config"
	"co-drive/pkg/database"
)

func main() {
	configPath := flag.String("config", "", "Path to configuration file (optional)")
	direction := flag.String("direction", "up", "Migration direction: up or down")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewPostgres(cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db, *direction); err != nil {
		log.Fatalf("Migration failed (%s): %v", *direction, err)
	}

	log.Printf("Successfully ran database migrations (%s)", *direction)
	os.Exit(0)
}
