package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCSRFProtect(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	protected := CSRFProtect(dummyHandler)

	t.Run("GET request sets CSRF cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		protected.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}

		cookies := rec.Result().Cookies()
		var csrfCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == "__csrf" {
				csrfCookie = c
				break
			}
		}

		if csrfCookie == nil {
			t.Fatal("Expected __csrf cookie to be set")
		}
		if len(csrfCookie.Value) != 64 {
			t.Errorf("Expected CSRF token length 64 (32 hex bytes), got %d", len(csrfCookie.Value))
		}
	})

	t.Run("POST request without token returns 403 Forbidden", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/submit", nil)
		rec := httptest.NewRecorder()

		protected.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", rec.Code)
		}
	})

	t.Run("POST request with valid form token passes", func(t *testing.T) {
		// First do GET to establish cookie
		getReq := httptest.NewRequest("GET", "/", nil)
		getRec := httptest.NewRecorder()
		protected.ServeHTTP(getRec, getReq)

		cookies := getRec.Result().Cookies()
		var csrfToken string
		for _, c := range cookies {
			if c.Name == "__csrf" {
				csrfToken = c.Value
				break
			}
		}

		// Perform POST with form token and cookie
		form := url.Values{}
		form.Set("csrf_token", csrfToken)
		postReq := httptest.NewRequest("POST", "/submit", strings.NewReader(form.Encode()))
		postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		postReq.AddCookie(&http.Cookie{Name: "__csrf", Value: csrfToken})
		postRec := httptest.NewRecorder()

		protected.ServeHTTP(postRec, postReq)

		if postRec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", postRec.Code)
		}
	})

	t.Run("POST request with valid header token passes", func(t *testing.T) {
		csrfToken := "11223344556677889900aabbccddeeff11223344556677889900aabbccddeeff"

		postReq := httptest.NewRequest("POST", "/api/action", nil)
		postReq.Header.Set("X-CSRF-Token", csrfToken)
		postReq.AddCookie(&http.Cookie{Name: "__csrf", Value: csrfToken})
		postRec := httptest.NewRecorder()

		protected.ServeHTTP(postRec, postReq)

		if postRec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", postRec.Code)
		}
	})

	t.Run("POST request with mismatched token returns 403 Forbidden", func(t *testing.T) {
		postReq := httptest.NewRequest("POST", "/api/action", nil)
		postReq.Header.Set("X-CSRF-Token", "wrongtoken")
		postReq.AddCookie(&http.Cookie{Name: "__csrf", Value: "11223344556677889900aabbccddeeff11223344556677889900aabbccddeeff"})
		postRec := httptest.NewRecorder()

		protected.ServeHTTP(postRec, postReq)

		if postRec.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", postRec.Code)
		}
	})
}
