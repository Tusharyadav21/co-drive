package expiry

import (
	"time"
)

const msPerDay = 24 * 60 * 60 * 1000

// ExpiryStatus represents the status of an expiring document
type ExpiryStatus string

const (
	StatusExpired  ExpiryStatus = "expired"
	StatusExpiring ExpiryStatus = "expiring"
	StatusActive   ExpiryStatus = "active"
	StatusNone     ExpiryStatus = "none"
)

const (
	PUCWarningDays       = 7
	InsuranceWarningDays = 15
)

// DaysUntilExpiry returns whole days from now until expiry (negative if expired)
func DaysUntilExpiry(expiryDate time.Time) int {
	if expiryDate.IsZero() {
		return 0
	}
	days := int(expiryDate.Sub(time.Now().Truncate(24 * time.Hour)).Hours() / 24)
	return days
}

// GetStatus returns the expiry status based on warning days
func GetStatus(expiryDate time.Time, warningDays int) ExpiryStatus {
	if expiryDate.IsZero() {
		return StatusNone
	}
	days := DaysUntilExpiry(expiryDate)
	if days < 0 {
		return StatusExpired
	}
	if days <= warningDays {
		return StatusExpiring
	}
	return StatusActive
}

// PUCStatus returns status for PUC (7-day warning)
func PUCStatus(expiryDate time.Time) ExpiryStatus {
	return GetStatus(expiryDate, PUCWarningDays)
}

// InsuranceStatus returns status for Insurance (15-day warning)
func InsuranceStatus(expiryDate time.Time) ExpiryStatus {
	return GetStatus(expiryDate, InsuranceWarningDays)
}

// GetStatusLabel returns a human-readable label for the status
func GetStatusLabel(status ExpiryStatus) string {
	switch status {
	case StatusExpired:
		return "Expired"
	case StatusExpiring:
		return "Expiring Soon"
	case StatusActive:
		return "Active"
	default:
		return "Missing"
	}
}

// GetStatusColor returns CSS color classes for the status
func GetStatusColor(status ExpiryStatus) string {
	switch status {
	case StatusExpired:
		return "text-red-500"
	case StatusExpiring:
		return "text-amber-500"
	case StatusActive:
		return "text-emerald-600 dark:text-emerald-400"
	default:
		return "text-zinc-400"
	}
}

// GetBadgeColor returns badge background color classes
func GetBadgeColor(status ExpiryStatus) string {
	switch status {
	case StatusExpired:
		return "bg-red-500/10 text-red-500 hover:bg-red-500/15"
	case StatusExpiring:
		return "bg-amber-500/10 text-amber-500 hover:bg-amber-500/15"
	case StatusActive:
		return "bg-emerald-500/10 text-emerald-500 hover:bg-emerald-500/15"
	default:
		return "bg-zinc-100 text-zinc-400 hover:bg-zinc-200 dark:bg-zinc-800 dark:hover:bg-zinc-700"
	}
}