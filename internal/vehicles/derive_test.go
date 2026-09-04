package vehicles

import (
	"testing"
	"time"
)

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestExpiry(t *testing.T) {
	now := day(2026, time.July, 29)

	tests := []struct {
		name         string
		expiry       time.Time
		wantExpired  bool
		wantDaysLeft int
	}{
		{"expires in 30 days", day(2026, time.August, 28), false, 30},
		{"expires today", now, false, 0},
		{"expired yesterday", day(2026, time.July, 28), true, -1},
		{"expired long ago", day(2025, time.July, 29), true, -365},
		{"across a leap day", day(2028, time.March, 1), false, 581}, // 2028 is a leap year
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Expiry(tt.expiry, now)
			if got.IsExpired != tt.wantExpired {
				t.Errorf("IsExpired = %v, want %v", got.IsExpired, tt.wantExpired)
			}
			if got.DaysRemaining != tt.wantDaysLeft {
				t.Errorf("DaysRemaining = %d, want %d", got.DaysRemaining, tt.wantDaysLeft)
			}
		})
	}
}

func litres(f float64) *float64 { return &f }

func TestApplyEfficiency(t *testing.T) {
	tests := []struct {
		name    string
		entries []MileageEntry
		want    []float64
	}{
		{
			name:    "first entry has no predecessor",
			entries: []MileageEntry{{Mileage: 1000, FuelLitres: litres(40)}},
			want:    []float64{0},
		},
		{
			name: "400km on 40L is 10 km/L",
			entries: []MileageEntry{
				{Mileage: 1000, FuelLitres: litres(40)},
				{Mileage: 1400, FuelLitres: litres(40)},
			},
			want: []float64{0, 10},
		},
		{
			name: "no fuel logged leaves efficiency at zero",
			entries: []MileageEntry{
				{Mileage: 1000},
				{Mileage: 1400},
			},
			want: []float64{0, 0},
		},
		{
			name: "zero litres does not divide by zero",
			entries: []MileageEntry{
				{Mileage: 1000, FuelLitres: litres(40)},
				{Mileage: 1400, FuelLitres: litres(0)},
			},
			want: []float64{0, 0},
		},
		{
			name: "odometer that does not advance stays at zero",
			entries: []MileageEntry{
				{Mileage: 1400, FuelLitres: litres(40)},
				{Mileage: 1400, FuelLitres: litres(40)},
			},
			want: []float64{0, 0},
		},
		{
			name: "odometer rollback stays at zero rather than going negative",
			entries: []MileageEntry{
				{Mileage: 1400, FuelLitres: litres(40)},
				{Mileage: 1000, FuelLitres: litres(40)},
			},
			want: []float64{0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ApplyEfficiency(tt.entries)
			for i, want := range tt.want {
				if tt.entries[i].Efficiency != want {
					t.Errorf("entry %d efficiency = %v, want %v", i, tt.entries[i].Efficiency, want)
				}
			}
		})
	}
}

func TestApplyEfficiencyEmpty(t *testing.T) {
	ApplyEfficiency(nil)              // must not panic
	ApplyEfficiency([]MileageEntry{}) // must not panic
}

func TestApplyDerivedTolerantOfMissingRecords(t *testing.T) {
	v := &Vehicle{Name: "no documents on file"}
	v.ApplyDerived(day(2026, time.July, 29)) // nil PUC, nil Insurance, nil mileage

	if v.PUC != nil || v.Insurance != nil {
		t.Fatal("ApplyDerived should not invent records")
	}
}

func TestApplyDerivedFillsNestedRecords(t *testing.T) {
	now := day(2026, time.July, 29)
	v := &Vehicle{
		PUC:       &VehicleDocument{Type: DocumentPUC, ExpiryDate: day(2026, time.July, 1)},
		Insurance: &VehicleDocument{Type: DocumentInsurance, ExpiryDate: day(2026, time.August, 28)},
		Mileage: []MileageEntry{
			{Mileage: 1000, FuelLitres: litres(40)},
			{Mileage: 1400, FuelLitres: litres(40)},
		},
	}

	v.ApplyDerived(now)

	if !v.PUC.IsExpired {
		t.Error("PUC expiring before now should be expired")
	}
	if v.Insurance.IsExpired || v.Insurance.DaysRemaining != 30 {
		t.Errorf("insurance = %+v, want valid with 30 days remaining", v.Insurance)
	}
	if v.Mileage[1].Efficiency != 10 {
		t.Errorf("mileage efficiency = %v, want 10", v.Mileage[1].Efficiency)
	}
}
