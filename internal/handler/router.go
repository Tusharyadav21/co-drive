package handler

import (
	"bytes"
	"embed"
	"html/template"
	"log/slog"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"co-drive/internal/auth"
	"co-drive/internal/db"
	"co-drive/internal/expiry"
	"co-drive/internal/middleware"
)

//go:embed templates/*.html templates/partials/*.html
var templatesFS embed.FS

type TemplateRenderer struct {
	pages    map[string]*template.Template
	partials *template.Template
}

var pageTemplates = []string{
	"login", "dashboard", "add_vehicle", "vehicle_detail",
	"add_mileage", "update_puc", "update_insurance", "profile",
}

func NewTemplateRenderer() (*TemplateRenderer, error) {
	funcMap := template.FuncMap{
		"len":            safeLen,
		"add":            func(a, b int) int { return a + b },
		"sub":            func(a, b int) int { return a - b },
		"formatMileage":  formatMileage,
		"formatCurrency": formatCurrency,
		"expiryStatus":  expiry.GetStatus,
		"firstLetters":  func(s string) string { if len(s) == 0 { return "?" }; return string([]rune(s)[0]) },
		"firstTwo":      func(s string) string { r := []rune(s); if len(r) == 0 { return "?" }; if len(r) == 1 { return string(r[0]) }; return string(r[:2]) },
		"upper":         strings.ToUpper,
		"daysUntil":     func(t time.Time) int { return expiry.DaysUntilExpiry(t) },
		"seq":           func(start, end int) []int { n := end - start + 1; if n < 0 { n = 0 }; s := make([]int, n); for i := range s { s[i] = start + i }; return s },
		"gridY":         func(i int) float64 { return float64(i)*18.8 + 5 },
		"dict": func(values ...interface{}) map[string]interface{} {
			if len(values)%2 != 0 {
				return nil
			}
			m := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil
				}
				m[key] = values[i+1]
			}
			return m
		},
	}

	// Parse partials (HTMX) separately for RenderPartial calls
	partials, err := template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/partials/*.html")
	if err != nil {
		return nil, err
	}

	// Each page gets its own template tree: layout + all partials + page-specific blocks
	// This avoids {{define "content"}} collisions across pages
	pages := make(map[string]*template.Template, len(pageTemplates))
	for _, name := range pageTemplates {
		t := template.New(name).Funcs(funcMap)
		t, err = t.ParseFS(templatesFS,
			"templates/layout.html",
			"templates/partials/*.html",
			"templates/"+name+".html",
		)
		if err != nil {
			return nil, err
		}
		pages[name] = t
	}

	return &TemplateRenderer{pages: pages, partials: partials}, nil
}

func (r *TemplateRenderer) Render(w http.ResponseWriter, name string, data interface{}) error {
	var buf bytes.Buffer
	var err error
	t, ok := r.pages[name]
	if !ok {
		err = r.partials.ExecuteTemplate(&buf, name, data)
	} else {
		err = t.ExecuteTemplate(&buf, "layout.html", data)
	}
	if err != nil {
		slog.Error("Template render error", "page", name, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
	return nil
}

func (r *TemplateRenderer) RenderPartial(w http.ResponseWriter, name string, data interface{}) error {
	var buf bytes.Buffer
	err := r.partials.ExecuteTemplate(&buf, name, data)
	if err != nil {
		slog.Error("Partial render error", "name", name, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
	return nil
}

func formatMileage(val interface{}) string {
	if val == nil {
		return "0"
	}
	var m int
	switch v := val.(type) {
	case int:
		m = v
	case int64:
		m = int(v)
	case float64:
		m = int(v)
	case float32:
		m = int(v)
	default:
		return "0"
	}
	if m == 0 {
		return "0"
	}
	s := ""
	for i := len(strconv.Itoa(m)); i > 0; i -= 3 {
		start := i - 3
		if start < 0 {
			start = 0
		}
		part := strconv.Itoa(m)[start:i]
		if s != "" {
			s = part + "," + s
		} else {
			s = part
		}
	}
	return s
}

func formatCurrency(val interface{}) string {
	if val == nil {
		return "0"
	}
	var f float64
	switch v := val.(type) {
	case float64:
		f = v
	case float32:
		f = float64(v)
	case int:
		f = float64(v)
	case int64:
		f = float64(v)
	default:
		return "0"
	}
	if f == 0 {
		return "0"
	}
	s := strconv.FormatFloat(f, 'f', 0, 64)
	parts := strings.Split(s, ".")
	result := ""
	for i := len(parts[0]) - 1; i >= 0; i-- {
		if (len(parts[0])-i)%3 == 0 && i != len(parts[0])-1 {
			result = "," + result
		}
		result = string(parts[0][i]) + result
	}
	return result
}

func safeLen(v interface{}) int {
	if v == nil {
		return 0
	}
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return val.Len()
	default:
		return 0
	}
}

type Services struct {
	Auth     *auth.Service
	Queries  *db.Queries
	Renderer *TemplateRenderer
}

type Handlers struct {
	*Services
}

func (h *Handlers) Render(w http.ResponseWriter, r *http.Request, name string, data interface{}) error {
	if m, ok := data.(map[string]interface{}); ok {
		if _, exists := m["CSRFToken"]; !exists {
			m["CSRFToken"] = middleware.GetCSRFToken(r)
		}
	}
	return h.Renderer.Render(w, name, data)
}

func (h *Handlers) RenderPartial(w http.ResponseWriter, r *http.Request, name string, data interface{}) error {
	if m, ok := data.(map[string]interface{}); ok {
		if _, exists := m["CSRFToken"]; !exists {
			m["CSRFToken"] = middleware.GetCSRFToken(r)
		}
	}
	return h.Renderer.RenderPartial(w, name, data)
}

func (h *Handlers) renderError(w http.ResponseWriter, msg string, status int) {
	slog.Error("Handler error", "status", status, "message", msg)
	http.Error(w, msg, status)
}

func New(pool *db.Pool, authSvc *auth.Service, _ ...interface{}) (*Handlers, error) {
	renderer, err := NewTemplateRenderer()
	if err != nil {
		return nil, err
	}

	services := &Services{
		Auth:     authSvc,
		Queries:  db.NewQueries(pool.Pool),
		Renderer: renderer,
	}

	return &Handlers{Services: services}, nil
}

func (h *Handlers) RegisterRoutes(mux *http.ServeMux) {
	// Auth routes (public)
	mux.HandleFunc("GET /auth/login", h.LoginPage)
	mux.HandleFunc("GET /auth/google", h.GoogleLogin)
	mux.HandleFunc("GET /auth/google/callback", h.GoogleCallback)
	mux.HandleFunc("POST /auth/logout", h.Logout)
	mux.HandleFunc("POST /auth/dev-login", h.DevLogin)

	// Protected routes
	mux.Handle("GET /{$}", middleware.RequireAuth(h.Auth, "/auth/login")(http.HandlerFunc(h.Dashboard)))
	mux.Handle("GET /vehicles/new", middleware.RequireAuth(h.Auth, "/auth/login")(http.HandlerFunc(h.NewVehiclePage)))
	mux.Handle("POST /vehicles/new", middleware.RequireAuth(h.Auth, "/auth/login")(http.HandlerFunc(h.CreateVehicle)))
	mux.Handle("GET /vehicles/{id}", middleware.RequireAuth(h.Auth, "/auth/login")(http.HandlerFunc(h.VehicleDetail)))
	mux.Handle("GET /vehicles/{id}/mileage", middleware.RequireAuth(h.Auth, "/auth/login")(http.HandlerFunc(h.AddMileagePage)))
	mux.Handle("POST /vehicles/{id}/mileage", middleware.RequireAuth(h.Auth, "/auth/login")(http.HandlerFunc(h.CreateMileage)))
	mux.Handle("GET /vehicles/{id}/puc", middleware.RequireAuth(h.Auth, "/auth/login")(http.HandlerFunc(h.PUCPage)))
	mux.Handle("POST /vehicles/{id}/puc", middleware.RequireAuth(h.Auth, "/auth/login")(http.HandlerFunc(h.UpdatePUC)))
	mux.Handle("GET /vehicles/{id}/insurance", middleware.RequireAuth(h.Auth, "/auth/login")(http.HandlerFunc(h.InsurancePage)))
	mux.Handle("POST /vehicles/{id}/insurance", middleware.RequireAuth(h.Auth, "/auth/login")(http.HandlerFunc(h.UpdateInsurance)))
	mux.Handle("GET /vehicles/{id}/tabs/{tab}", middleware.RequireAuth(h.Auth, "/auth/login")(http.HandlerFunc(h.VehicleTab)))
	mux.Handle("POST /vehicles/{id}/share", middleware.RequireAuth(h.Auth, "/auth/login")(http.HandlerFunc(h.ShareVehicle)))
	mux.Handle("POST /vehicles/{id}/unshare/{vuID}", middleware.RequireAuth(h.Auth, "/auth/login")(http.HandlerFunc(h.UnshareVehicle)))
	mux.Handle("GET /profile", middleware.RequireAuth(h.Auth, "/auth/login")(http.HandlerFunc(h.Profile)))

	// API routes
	mux.Handle("POST /api/notifications/subscribe", middleware.RequireAuth(h.Auth, "/auth/login")(http.HandlerFunc(h.SubscribePush)))
}