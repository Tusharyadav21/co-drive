package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"co-drive/config"
	"co-drive/internal/auth"
	"co-drive/internal/middleware"
	"co-drive/internal/users"
	"co-drive/internal/vehicles"
	"co-drive/pkg/database"
	"co-drive/pkg/email"
)

func main() {
	configPath := flag.String("config", "", "Path to config file (optional)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 1. Initialize PostgreSQL Database
	db, err := database.NewPostgres(cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Auto-run embedded migrations up
	if err := database.RunMigrations(db, "up"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// 2. Initialize Repositories
	userRepo := users.NewRepository(db)
	sessionRepo := auth.NewSessionRepository(db)
	verificationRepo := auth.NewVerificationRepository(db)
	vehicleRepo := vehicles.NewRepository(db)

	// 3. Initialize Services
	emailService := email.NewSMTPService(cfg.SMTP)
	sessionService := auth.NewSessionService(sessionRepo)
	verificationService := auth.NewVerificationService(verificationRepo, emailService, cfg.Auth.OTPTTLMinutes, cfg.Auth.PasswordResetTTLMinutes)

	// 4. Initialize Handlers & Middlewares
	authHandler := auth.NewHandler(userRepo, sessionService)
	verificationHandler := auth.NewVerificationHandler(verificationService, userRepo, sessionService)
	userHandler := users.NewHandler(userRepo)
	vehicleHandler := vehicles.NewHandler(vehicleRepo)

	authMiddleware := middleware.NewAuth(sessionService)
	csrfMiddleware := middleware.NewCSRF()
	rateLimiter := middleware.NewRateLimiter(100, 1*time.Minute)

	// 5. Router Setup
	workDir, err := os.Getwd()
	if err != nil {
		workDir = "."
	}

	r := NewRouter(RouterDeps{
		AuthHandler:         authHandler,
		VerificationHandler: verificationHandler,
		UserHandler:         userHandler,
		VehicleHandler:      vehicleHandler,
		AuthMiddleware:      authMiddleware,
		CSRFMiddleware:      csrfMiddleware,
		RateLimiter:         rateLimiter,
		StaticDir:           filepath.Join(workDir, "static"),
	})

	// 6. Start HTTP Server with Graceful Shutdown
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server ListenAndServe failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced shutdown: %v", err)
	}

	log.Println("Server stopped cleanly.")
}
