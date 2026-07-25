package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTemplateRenderer(t *testing.T) {
	renderer, err := NewTemplateRenderer()
	if err != nil {
		t.Fatalf("Failed to create template renderer: %v", err)
	}

	pages := []string{"login", "dashboard", "add_vehicle", "vehicle_detail", "add_mileage", "update_puc", "update_insurance", "profile"}
	for _, page := range pages {
		rec := httptest.NewRecorder()
		err := renderer.Render(rec, page, map[string]interface{}{
			"Title": "Test",
			"User":  map[string]interface{}{"Name": "Test User"},
		})
		if err != nil {
			t.Errorf("Failed to render page %s: %v", page, err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("Page %s returned status %d", page, rec.Code)
		}
	}
}
