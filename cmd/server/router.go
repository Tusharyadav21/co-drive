package main

import (
	"net/http"
	"path/filepath"

	"co-drive/internal/auth"
	"co-drive/internal/middleware"
	"co-drive/internal/users"
	"co-drive/internal/vehicles"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

// RouterDeps is everything the routing table needs. Keeping it a plain struct is
// what lets a test build a router over fakes and assert the table with httptest.
type RouterDeps struct {
	AuthHandler         *auth.Handler
	VerificationHandler *auth.VerificationHandler
	UserHandler         *users.Handler
	VehicleHandler      *vehicles.Handler

	AuthMiddleware *middleware.Auth
	CSRFMiddleware *middleware.CSRF
	RateLimiter    *middleware.RateLimiter

	StaticDir string
}

// staticPages maps a URL path to the file under StaticDir that serves it. Both
// the bare and trailing-slash forms are registered for each.
var staticPages = map[string]string{
	"/":          "index.html",
	"/auth":      "auth.html",
	"/dashboard": "dashboard.html",
	"/app":       "dashboard.html",
	"/analytics": "analytics.html",
	"/profile":   "profile.html",
}

func NewRouter(d RouterDeps) http.Handler {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.Logging)
	r.Use(middleware.SecurityHeaders)
	r.Use(d.RateLimiter.Middleware)
	r.Use(d.CSRFMiddleware.Middleware)

	// Public Auth Routes
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", d.AuthHandler.Register)
		r.Post("/login", d.AuthHandler.Login)

		// Email OTP passwordless login / 2FA
		r.Post("/otp/request", d.VerificationHandler.RequestOTP)
		r.Post("/otp/verify", d.VerificationHandler.VerifyOTP)
	})

	// Protected Routes
	r.Group(func(r chi.Router) {
		r.Use(d.AuthMiddleware.Authenticate)

		r.Post("/api/auth/logout", d.AuthHandler.Logout)

		// User profile & details routes
		r.Get("/api/users/me", d.UserHandler.GetProfile)
		r.Put("/api/users/me", d.UserHandler.UpdateProfile)

		// Vehicle Management API Routes
		r.Route("/api/vehicles", func(r chi.Router) {
			r.Get("/", d.VehicleHandler.ListVehicles)
			r.Post("/", d.VehicleHandler.CreateVehicle)
			r.Get("/{id}", d.VehicleHandler.GetVehicle)
			r.Put("/{id}", d.VehicleHandler.UpdateVehicle)
			r.Delete("/{id}", d.VehicleHandler.DeleteVehicle)
			r.Post("/{id}/mileage", d.VehicleHandler.AddMileage)
			r.Put("/{id}/mileage/{mileageId}", d.VehicleHandler.UpdateMileage)
			r.Delete("/{id}/mileage/{mileageId}", d.VehicleHandler.DeleteMileage)
			r.Post("/{id}/puc", d.VehicleHandler.UpdatePUC)
			r.Post("/{id}/insurance", d.VehicleHandler.UpdateInsurance)
		})
	})

	// Serve static frontend UI pages & assets
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(d.StaticDir))))

	for path, file := range staticPages {
		serve := servePage(filepath.Join(d.StaticDir, file))
		r.Get(path, serve)
		if path != "/" {
			r.Get(path+"/", serve)
		}
	}

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	return r
}

func servePage(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		http.ServeFile(w, r, path)
	}
}
