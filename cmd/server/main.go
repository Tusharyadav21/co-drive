package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"co-drive/internal/auth"
	"co-drive/internal/db"
	"co-drive/internal/handler"
	"co-drive/internal/middleware"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Load configuration
	cfg := loadConfig()

	// Setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Generate session secret if not provided
	if cfg.SessionSecret == "" {
		cfg.SessionSecret = generateSecret()
		slog.Info("Generated session secret", "length", len(cfg.SessionSecret))
	}

	// Database connection
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Run migrations
	if err := db.RunMigrations(ctx, pool); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Initialize auth
	authSvc := auth.NewService(pool.Pool, cfg.Config)

	// Initialize handlers
	h, err := handler.New(pool, authSvc)
	if err != nil {
		slog.Error("Failed to create handlers", "error", err)
		os.Exit(1)
	}

	// Middleware chain
	mux := http.NewServeMux()

	// Static files (public)
	execDir, _ := os.Getwd()
	staticDir := filepath.Join(execDir, "static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	h.RegisterRoutes(mux)

	handler := middleware.SecurityHeaders(
		middleware.RequestLogger(
			middleware.AuthMiddleware(authSvc,
				middleware.CSRFProtect(
					mux,
				),
			),
		),
	)

	// HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server
	go func() {
		slog.Info("Starting server", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	// Start background expiry check
	go authSvc.StartExpiryChecker(ctx)

	// Wait for shutdown signal
	<-ctx.Done()
	slog.Info("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server shutdown error", "error", err)
	}

	slog.Info("Server stopped")
}

type config struct {
	DatabaseURL   string
	Port          string
	SessionSecret string
	auth.Config
}

func loadConfig() config {
	return config{
		DatabaseURL:   getEnv("DATABASE_URL", ""),
		Port:          getEnv("PORT", "8080"),
		SessionSecret: getEnv("SESSION_SECRET", ""),
		Config: auth.Config{
			CookieDomain:    getEnv("COOKIE_DOMAIN", ""),
			GoogleClientID:  getEnv("GOOGLE_CLIENT_ID", ""),
			GoogleSecret:    getEnv("GOOGLE_CLIENT_SECRET", ""),
			GoogleRedirect:  getEnv("GOOGLE_REDIRECT_URL", ""),
			VAPIDPublicKey:  getEnv("VAPID_PUBLIC_KEY", ""),
			VAPIDPrivateKey: getEnv("VAPID_PRIVATE_KEY", ""),
			VAPIDSubject:    getEnv("VAPID_SUBJECT", "mailto:admin@codrive.app"),
			DevLoginEnabled: getEnv("DEV_LOGIN_ENABLED", "false") == "true",
		},
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func generateSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}