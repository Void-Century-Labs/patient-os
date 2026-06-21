package main

import (
	"log"

	"github.com/void-century-labs/patient-os/backend/internal/config"
	"github.com/void-century-labs/patient-os/backend/internal/database"
	"github.com/void-century-labs/patient-os/backend/internal/router"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	r := router.New(db)

	log.Printf("PatientOS backend listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
