package vehicles

import "time"

type Vehicle struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Name         string    `json:"name"`
	Make         string    `json:"make"`
	Model        string    `json:"model"`
	Year         int       `json:"year"`
	LicensePlate     string    `json:"license_plate"`
	ChassisNumber    string    `json:"chassis_number,omitempty"`
	EngineNumber     string    `json:"engine_number,omitempty"`
	RegistrationDate string    `json:"registration_date,omitempty"`
	NextServiceKM    int       `json:"next_service_km,omitempty"`
	NextServiceDate  string    `json:"next_service_date,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// Enriched fields for API responses
	PUC       *VehicleDocument `json:"puc,omitempty"`
	Insurance *VehicleDocument `json:"insurance,omitempty"`
	Mileage   []MileageEntry   `json:"mileage_entries,omitempty"`
}

type MileageEntry struct {
	ID         string    `json:"id"`
	VehicleID  string    `json:"vehicle_id"`
	Mileage    int       `json:"mileage"`
	Date       time.Time `json:"date"`
	Notes      string    `json:"notes"`
	FuelLitres *float64  `json:"fuel_litres,omitempty"`
	FuelAmount *float64  `json:"fuel_amount,omitempty"`
	Efficiency float64   `json:"efficiency,omitempty"` // calculated km/L
	CreatedAt  time.Time `json:"created_at"`
}

// DocumentType distinguishes the rows of the single vehicle_documents table.
// PUC and insurance were identical shapes under different column names; this
// is that shape, generalized. A future document type (road tax, fitness
// certificate) is a new constant and a row, not a migration.
type DocumentType string

const (
	DocumentPUC       DocumentType = "puc"
	DocumentInsurance DocumentType = "insurance"
)

type VehicleDocument struct {
	ID            string       `json:"id"`
	VehicleID     string       `json:"vehicle_id"`
	Type          DocumentType `json:"type"`
	Reference     string       `json:"reference"` // certificate number (PUC) or policy number (insurance)
	Issuer        string       `json:"issuer,omitempty"`
	ExpiryDate    time.Time    `json:"expiry_date"`
	IsExpired     bool         `json:"is_expired"`
	DaysRemaining int          `json:"days_remaining"`
	UpdatedAt     time.Time    `json:"updated_at"`
}
