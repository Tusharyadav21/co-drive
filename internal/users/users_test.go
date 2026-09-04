package users

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"co-drive/internal/session"
	"co-drive/pkg/database"
	"github.com/google/uuid"
)

func TestUsersRepositoryAndProfile(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// 1. Create user with blank full name -> should default to "Driver"
	u1 := &User{
		ID:    uuid.NewString(),
		Email: "driver1@example.com",
	}
	if err := repo.Create(ctx, u1); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	fetched1, err := repo.GetByID(ctx, u1.ID)
	if err != nil || fetched1 == nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if fetched1.FullName != "Driver" {
		t.Errorf("expected default full_name 'Driver', got %q", fetched1.FullName)
	}

	// 2. Create user with specified full name
	u2 := &User{
		ID:       uuid.NewString(),
		Email:    "driver2@example.com",
		FullName: "Jane Doe",
	}
	if err := repo.Create(ctx, u2); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	fetched2, err := repo.GetByEmail(ctx, "driver2@example.com")
	if err != nil || fetched2 == nil {
		t.Fatalf("failed to get user by email: %v", err)
	}
	if fetched2.FullName != "Jane Doe" {
		t.Errorf("expected full_name 'Jane Doe', got %q", fetched2.FullName)
	}

	// 3. Test initial GetProfile
	prof, err := repo.GetProfile(ctx, u1.ID)
	if err != nil || prof == nil {
		t.Fatalf("failed to get profile: %v", err)
	}
	if prof.Email != "driver1@example.com" {
		t.Errorf("expected email 'driver1@example.com', got %q", prof.Email)
	}
	if prof.FullName != "Driver" {
		t.Errorf("expected full_name 'Driver', got %q", prof.FullName)
	}
	if prof.Phone != "" || prof.Bio != "" || prof.Address != "" {
		t.Errorf("expected empty details for initial profile, got phone=%q, bio=%q, addr=%q", prof.Phone, prof.Bio, prof.Address)
	}

	// 4. Update Profile
	updateReq := &UpdateProfileRequest{
		FullName:         "Alex Mercer",
		Phone:            "+1-555-0199",
		Bio:              "Fleet enthusiast and daily commuter.",
		Address:          "742 Evergreen Terrace",
		EmergencyContact: "+1-555-0911",
		AvatarURL:        "https://example.com/avatar.jpg",
	}
	updatedProf, err := repo.UpdateProfile(ctx, u1.ID, updateReq)
	if err != nil {
		t.Fatalf("failed to update profile: %v", err)
	}
	if updatedProf.FullName != "Alex Mercer" {
		t.Errorf("expected updated full_name 'Alex Mercer', got %q", updatedProf.FullName)
	}
	if updatedProf.Phone != updateReq.Phone || updatedProf.Bio != updateReq.Bio || updatedProf.Address != updateReq.Address || updatedProf.EmergencyContact != updateReq.EmergencyContact || updatedProf.AvatarURL != updateReq.AvatarURL {
		t.Errorf("profile fields mismatch after update: %+v", updatedProf)
	}

	// 5. Subsequent GetProfile reflects updates
	profRefreshed, err := repo.GetProfile(ctx, u1.ID)
	if err != nil || profRefreshed == nil {
		t.Fatalf("failed to get refreshed profile: %v", err)
	}
	if profRefreshed.FullName != "Alex Mercer" || profRefreshed.Phone != updateReq.Phone {
		t.Errorf("refreshed profile mismatch: %+v", profRefreshed)
	}

	// 6. Test UpdateEmailVerified
	if err := repo.UpdateEmailVerified(ctx, u1.ID, true); err != nil {
		t.Fatalf("failed to update email verified: %v", err)
	}
	profVerified, _ := repo.GetProfile(ctx, u1.ID)
	if !profVerified.EmailVerified {
		t.Errorf("expected email_verified = true")
	}
}

func TestUsersHTTPHandler(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	handler := NewHandler(repo)
	ctx := context.Background()

	testUser := &User{
		ID:       uuid.NewString(),
		Email:    "testuser@example.com",
		FullName: "Test User",
	}
	if err := repo.Create(ctx, testUser); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Case 1: GetProfile without session -> 401
	reqUnauth := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	recUnauth := httptest.NewRecorder()
	handler.GetProfile(recUnauth, reqUnauth)
	if recUnauth.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", recUnauth.Code)
	}

	// Case 2: GetProfile with session context -> 200
	reqAuth := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	reqAuth = reqAuth.WithContext(session.WithSession(reqAuth.Context(), &session.Session{UserID: testUser.ID}))
	recAuth := httptest.NewRecorder()
	handler.GetProfile(recAuth, reqAuth)
	if recAuth.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", recAuth.Code, recAuth.Body.String())
	}

	var pResp ProfileResponse
	if err := json.NewDecoder(recAuth.Body).Decode(&pResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if pResp.ID != testUser.ID || pResp.FullName != "Test User" {
		t.Errorf("unexpected profile response: %+v", pResp)
	}

	// Case 3: UpdateProfile with bad JSON -> 400
	reqBadJSON := httptest.NewRequest(http.MethodPut, "/api/users/me", bytes.NewBufferString("{invalid-json"))
	reqBadJSON = reqBadJSON.WithContext(session.WithSession(reqBadJSON.Context(), &session.Session{UserID: testUser.ID}))
	recBadJSON := httptest.NewRecorder()
	handler.UpdateProfile(recBadJSON, reqBadJSON)
	if recBadJSON.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", recBadJSON.Code)
	}

	// Case 4: UpdateProfile valid payload -> 200
	updatePayload := UpdateProfileRequest{
		FullName: "Updated Name",
		Phone:    "+1-555-1234",
		Bio:      "Driver Bio",
	}
	bodyBytes, _ := json.Marshal(updatePayload)
	reqValidUpdate := httptest.NewRequest(http.MethodPut, "/api/users/me", bytes.NewBuffer(bodyBytes))
	reqValidUpdate = reqValidUpdate.WithContext(session.WithSession(reqValidUpdate.Context(), &session.Session{UserID: testUser.ID}))
	recValidUpdate := httptest.NewRecorder()
	handler.UpdateProfile(recValidUpdate, reqValidUpdate)
	if recValidUpdate.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", recValidUpdate.Code, recValidUpdate.Body.String())
	}

	var updatedResp ProfileResponse
	if err := json.NewDecoder(recValidUpdate.Body).Decode(&updatedResp); err != nil {
		t.Fatalf("failed to decode update response: %v", err)
	}
	if updatedResp.FullName != "Updated Name" || updatedResp.Phone != "+1-555-1234" || updatedResp.Bio != "Driver Bio" {
		t.Errorf("unexpected updated profile: %+v", updatedResp)
	}
}
