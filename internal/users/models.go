package users

import "time"

// User represents the core user account identity.
type User struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	FullName      string    `json:"full_name"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UserDetails holds extended profile metadata stored in the user_details table.
type UserDetails struct {
	UserID           string    `json:"user_id"`
	Phone            string    `json:"phone"`
	Bio              string    `json:"bio"`
	Address          string    `json:"address"`
	EmergencyContact string    `json:"emergency_contact"`
	AvatarURL        string    `json:"avatar_url"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ProfileResponse represents the consolidated user and details payload.
type ProfileResponse struct {
	ID               string    `json:"id"`
	Email            string    `json:"email"`
	FullName         string    `json:"full_name"`
	EmailVerified    bool      `json:"email_verified"`
	Phone            string    `json:"phone"`
	Bio              string    `json:"bio"`
	Address          string    `json:"address"`
	EmergencyContact string    `json:"emergency_contact"`
	AvatarURL        string    `json:"avatar_url"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// UpdateProfileRequest defines the payload for editing profile details.
type UpdateProfileRequest struct {
	FullName         string `json:"full_name"`
	Phone            string `json:"phone"`
	Bio              string `json:"bio"`
	Address          string `json:"address"`
	EmergencyContact string `json:"emergency_contact"`
	AvatarURL        string `json:"avatar_url"`
}
