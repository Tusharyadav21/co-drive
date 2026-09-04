package vehicles

import "time"

// The rules in this file are the part of the Vehicle Garage that can actually
// be wrong: how close a document is to expiry, and what fuel economy a pair of
// odometer readings implies. They take `now` as an argument rather than calling
// time.Now() so they can be tested without a clock or a database.

// ExpiryStatus is the derived view of any dated document (PUC, insurance).
type ExpiryStatus struct {
	IsExpired     bool
	DaysRemaining int
}

func Expiry(expiryDate, now time.Time) ExpiryStatus {
	return ExpiryStatus{
		IsExpired:     now.After(expiryDate),
		DaysRemaining: int(expiryDate.Sub(now).Hours() / 24),
	}
}

func (d *VehicleDocument) ApplyStatus(now time.Time) {
	if d == nil {
		return
	}
	s := Expiry(d.ExpiryDate, now)
	d.IsExpired, d.DaysRemaining = s.IsExpired, s.DaysRemaining
}

// ApplyEfficiency fills in km/L for each entry from the distance since the
// previous one. Entries must be in ascending date order. The first entry has no
// predecessor, and an entry with no fuel logged or no distance travelled keeps
// an efficiency of zero — there is nothing to divide.
func ApplyEfficiency(entries []MileageEntry) {
	for i := 1; i < len(entries); i++ {
		if entries[i].FuelLitres == nil || *entries[i].FuelLitres <= 0 {
			continue
		}
		dist := entries[i].Mileage - entries[i-1].Mileage
		if dist > 0 {
			entries[i].Efficiency = float64(dist) / *entries[i].FuelLitres
		}
	}
}

// ApplyDerived enriches a Vehicle for an API response.
func (v *Vehicle) ApplyDerived(now time.Time) {
	if v == nil {
		return
	}
	v.PUC.ApplyStatus(now)
	v.Insurance.ApplyStatus(now)
	ApplyEfficiency(v.Mileage)
}
