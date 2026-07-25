package handler

import (
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"co-drive/internal/db"
	"co-drive/internal/expiry"
	"co-drive/internal/middleware"
	"co-drive/internal/svg"
)

func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	vehicles, err := h.Queries.GetVehiclesByUserID(r.Context(), user.ID)
	if err != nil {
		slog.Error("Failed to get vehicles by user ID", "userID", user.ID, "error", err)
		h.renderError(w, "Failed to load vehicles", http.StatusInternalServerError)
		return
	}

	// Enrich with latest data
	type VehicleCard struct {
		db.Vehicle
		LatestMileage   *db.MileageEntry
		PUC             *db.PUC
		Insurance       *db.Insurance
		PUCStatus       string
		InsuranceStatus string
		IsTwoWheeler    bool
		Users           []db.VehicleUser
	}

	var cards []VehicleCard
	for _, v := range vehicles {
		details, _ := h.Queries.GetVehicleByID(r.Context(), v.ID, user.ID)
		var latestMileage *db.MileageEntry
		if len(details.MileageEntries) > 0 {
			latestMileage = &details.MileageEntries[0]
		}
		pucStatus := "none"
		if details.PUC != nil {
			pucStatus = string(expiry.GetStatus(details.PUC.ExpiryDate, 7))
		}
		insStatus := "none"
		if details.Insurance != nil {
			insStatus = string(expiry.GetStatus(details.Insurance.ExpiryDate, 15))
		}
		cards = append(cards, VehicleCard{
			Vehicle:         v,
			LatestMileage:   latestMileage,
			PUC:             details.PUC,
			Insurance:       details.Insurance,
			PUCStatus:       pucStatus,
			InsuranceStatus: insStatus,
			IsTwoWheeler:    isTwoWheeler(v),
			Users:           details.Users,
		})
	}

	stats, _ := h.Queries.GetUserStats(r.Context(), user.ID)

	activeAlerts := 0
	for _, c := range cards {
		if c.PUCStatus == "expiring" || c.PUCStatus == "expired" || c.InsuranceStatus == "expiring" || c.InsuranceStatus == "expired" {
			activeAlerts++
		}
	}

	data := map[string]interface{}{
		"Title":        "Dashboard",
		"User":         user,
		"Vehicles":     cards,
		"Stats":        stats,
		"ActiveAlerts": activeAlerts,
		"CurrentPath":  r.URL.Path,
	}
	h.Render(w, r, "dashboard", data)
}

func (h *Handlers) NewVehiclePage(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Add Vehicle",
		"MaxYear": time.Now().Year() + 1,
	}
	h.Render(w, r, "add_vehicle", data)
}

func (h *Handlers) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		h.renderError(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	make := r.FormValue("make")
	model := r.FormValue("model")
	yearStr := r.FormValue("year")
	licensePlate := r.FormValue("license_plate")

	if name == "" || make == "" || model == "" || yearStr == "" || licensePlate == "" {
		h.renderError(w, "All fields are required", http.StatusBadRequest)
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 1900 || year > time.Now().Year()+1 {
		h.renderError(w, "Invalid year", http.StatusBadRequest)
		return
	}

	vehicle, err := h.Queries.CreateVehicle(r.Context(), name, make, model, year, licensePlate, user.ID)
	if err != nil {
		h.renderError(w, "Failed to create vehicle", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/vehicles/"+vehicle.ID, http.StatusSeeOther)
}

func (h *Handlers) VehicleDetail(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	vehicleID := r.PathValue("id")
	if vehicleID == "" {
		h.renderError(w, "Vehicle not found", http.StatusNotFound)
		return
	}

	details, err := h.Queries.GetVehicleByID(r.Context(), vehicleID, user.ID)
	if err != nil {
		h.renderError(w, "Vehicle not found", http.StatusNotFound)
		return
	}

	enriched := enrichVehicleData(details)

	data := map[string]interface{}{
		"Vehicle":   enriched,
		"User":      user,
		"ActiveTab": "mileage",
		"Now":       time.Now(),
	}
	h.Render(w, r, "vehicle_detail", data)
}

type enrichedVehicle struct {
	db.VehicleWithDetails
	ChartData            []svg.ChartPoint
	ChartPoints          string
	AverageEfficiency    float64
	TotalTrackedDistance int
	IsTwoWheeler         bool
}

func enrichVehicleData(details *db.VehicleWithDetails) *enrichedVehicle {
	ev := &enrichedVehicle{
		VehicleWithDetails: *details,
		IsTwoWheeler:       isTwoWheeler(details.Vehicle),
	}

	if len(details.MileageEntries) < 2 {
		return ev
	}

	// Sort entries by date ascending for chart
	sorted := make([]db.MileageEntry, len(details.MileageEntries))
	copy(sorted, details.MileageEntries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Date.Before(sorted[j].Date)
	})

	// Build chart data and compute stats
	var chartData []svg.ChartPoint
	var totalFuelLitres float64

	for i, e := range sorted {
		chartData = append(chartData, svg.ChartPoint{
			Date:     e.Date.Format("02 Jan"),
			Odometer: float64(e.Mileage),
			Expense:  e.FuelAmount.Float64,
		})
		if e.FuelLitres.Valid {
			totalFuelLitres += e.FuelLitres.Float64
		}
		// Compute per-entry efficiency (km/l)
		if e.FuelLitres.Valid && e.FuelLitres.Float64 > 0 && i > 0 {
			dist := e.Mileage - sorted[i-1].Mileage
			if dist > 0 {
				sorted[i].Efficiency = float64(dist) / e.FuelLitres.Float64
			}
		}
	}

	ev.ChartData = chartData
	ev.ChartPoints = svg.FormatChartPoints(chartData)

	if totalFuelLitres > 0 {
		ev.AverageEfficiency = float64(sorted[len(sorted)-1].Mileage-sorted[0].Mileage) / totalFuelLitres
	}

	if len(sorted) >= 2 {
		ev.TotalTrackedDistance = sorted[len(sorted)-1].Mileage - sorted[0].Mileage
	}

	// Write computed efficiencies back to the original entries slice
	for i := range details.MileageEntries {
		for _, s := range sorted {
			if details.MileageEntries[i].ID == s.ID {
				details.MileageEntries[i].Efficiency = s.Efficiency
				break
			}
		}
	}

	return ev
}

func (h *Handlers) AddMileagePage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	vehicleID := r.PathValue("id")
	if vehicleID == "" {
		h.renderError(w, "Vehicle not found", http.StatusNotFound)
		return
	}

	data := map[string]interface{}{
		"VehicleID": vehicleID,
		"Today":     time.Now().Format("2006-01-02"),
	}
	h.Render(w, r, "add_mileage", data)
}

func (h *Handlers) CreateMileage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	vehicleID := r.PathValue("id")
	if vehicleID == "" {
		h.renderError(w, "Vehicle not found", http.StatusNotFound)
		return
	}

	hasAccess, _, _ := h.Queries.HasAccess(r.Context(), vehicleID, user.ID)
	if !hasAccess {
		h.renderError(w, "Access denied", http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		h.renderError(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	mileageStr := r.FormValue("mileage")
	dateStr := r.FormValue("date")
	notes := r.FormValue("notes")
	fuelLitresStr := r.FormValue("fuel_litres")
	fuelAmountStr := r.FormValue("fuel_amount")

	if mileageStr == "" || dateStr == "" {
		h.renderError(w, "Mileage and date are required", http.StatusBadRequest)
		return
	}

	mileage, err := strconv.Atoi(mileageStr)
	if err != nil || mileage < 0 {
		h.renderError(w, "Invalid mileage", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.renderError(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	var fuelLitres, fuelAmount *float64
	if fuelLitresStr != "" {
		if fl, err := strconv.ParseFloat(fuelLitresStr, 64); err == nil {
			fuelLitres = &fl
		}
	}
	if fuelAmountStr != "" {
		if fa, err := strconv.ParseFloat(fuelAmountStr, 64); err == nil {
			fuelAmount = &fa
		}
	}

	_, err = h.Queries.AddMileageEntry(r.Context(), vehicleID, user.ID, mileage, date, notes, fuelLitres, fuelAmount)
	if err != nil {
		h.renderError(w, "Failed to add mileage entry", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/vehicles/"+vehicleID, http.StatusSeeOther)
}

func (h *Handlers) PUCPage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	vehicleID := r.PathValue("id")
	if vehicleID == "" {
		h.renderError(w, "Vehicle not found", http.StatusNotFound)
		return
	}

	hasAccess, _, _ := h.Queries.HasAccess(r.Context(), vehicleID, user.ID)
	if !hasAccess {
		h.renderError(w, "Access denied", http.StatusForbidden)
		return
	}

	puc, _ := h.Queries.GetPUC(r.Context(), vehicleID)

	data := map[string]interface{}{
		"VehicleID": vehicleID,
		"PUC":       puc,
	}
	h.Render(w, r, "update_puc", data)
}

func (h *Handlers) UpdatePUC(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	vehicleID := r.PathValue("id")
	if vehicleID == "" {
		h.renderError(w, "Vehicle not found", http.StatusNotFound)
		return
	}

	isOwner, _ := h.Queries.IsOwner(r.Context(), vehicleID, user.ID)
	if !isOwner {
		h.renderError(w, "Only the vehicle owner can update PUC", http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		h.renderError(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	expiryStr := r.FormValue("expiry_date")
	certificateNumber := r.FormValue("certificate_number")

	if expiryStr == "" {
		h.renderError(w, "Expiry date is required", http.StatusBadRequest)
		return
	}

	expiry, err := time.Parse("2006-01-02", expiryStr)
	if err != nil {
		h.renderError(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	_, err = h.Queries.UpsertPUC(r.Context(), vehicleID, expiry, certificateNumber)
	if err != nil {
		h.renderError(w, "Failed to update PUC", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/vehicles/"+vehicleID, http.StatusSeeOther)
}

func (h *Handlers) InsurancePage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	vehicleID := r.PathValue("id")
	if vehicleID == "" {
		h.renderError(w, "Vehicle not found", http.StatusNotFound)
		return
	}

	hasAccess, _, _ := h.Queries.HasAccess(r.Context(), vehicleID, user.ID)
	if !hasAccess {
		h.renderError(w, "Access denied", http.StatusForbidden)
		return
	}

	insurance, _ := h.Queries.GetInsurance(r.Context(), vehicleID)

	data := map[string]interface{}{
		"VehicleID": vehicleID,
		"Insurance": insurance,
	}
	h.Render(w, r, "update_insurance", data)
}

func (h *Handlers) UpdateInsurance(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	vehicleID := r.PathValue("id")
	if vehicleID == "" {
		h.renderError(w, "Vehicle not found", http.StatusNotFound)
		return
	}

	isOwner, _ := h.Queries.IsOwner(r.Context(), vehicleID, user.ID)
	if !isOwner {
		h.renderError(w, "Only the vehicle owner can update Insurance", http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		h.renderError(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	expiryStr := r.FormValue("expiry_date")
	policyNumber := r.FormValue("policy_number")
	provider := r.FormValue("provider")

	if expiryStr == "" {
		h.renderError(w, "Expiry date is required", http.StatusBadRequest)
		return
	}

	expiry, err := time.Parse("2006-01-02", expiryStr)
	if err != nil {
		h.renderError(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	_, err = h.Queries.UpsertInsurance(r.Context(), vehicleID, expiry, policyNumber, provider)
	if err != nil {
		h.renderError(w, "Failed to update Insurance", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/vehicles/"+vehicleID, http.StatusSeeOther)
}

func isTwoWheeler(v db.Vehicle) bool {
	name := strings.ToLower(v.Name)
	make := strings.ToLower(v.Make)
	model := strings.ToLower(v.Model)
	twoWheelerKeywords := []string{"bike", "scooter", "motorcycle", "moped", "two-wheeler", "2-wheeler"}
	for _, kw := range twoWheelerKeywords {
		if strings.Contains(name, kw) || strings.Contains(make, kw) || strings.Contains(model, kw) {
			return true
		}
	}
	return false
}