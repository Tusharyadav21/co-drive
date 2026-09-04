package vehicles

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"co-drive/internal/session"
	"co-drive/pkg/response"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

type CreateVehicleRequest struct {
	Name         string `json:"name"`
	Make         string `json:"make"`
	Model        string `json:"model"`
	Year         int    `json:"year"`
	LicensePlate string `json:"license_plate"`
}

func (h *Handler) ListVehicles(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.Require(w, r)
	if !ok {
		return
	}

	list, err := h.repo.ListVehiclesByUserID(r.Context(), sess.UserID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to retrieve vehicles")
		return
	}

	now := time.Now()
	for i := range list {
		list[i].ApplyDerived(now)
	}

	response.JSON(w, http.StatusOK, list)
}

func (h *Handler) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.Require(w, r)
	if !ok {
		return
	}

	var req CreateVehicleRequest
	if err := response.DecodeJSON(r, &req); err != nil || req.Name == "" {
		response.Error(w, http.StatusBadRequest, "Vehicle name is required")
		return
	}

	if req.Year <= 0 {
		req.Year = time.Now().Year()
	}

	v := &Vehicle{
		UserID:       sess.UserID,
		Name:         req.Name,
		Make:         req.Make,
		Model:        req.Model,
		Year:         req.Year,
		LicensePlate: req.LicensePlate,
	}

	if err := h.repo.CreateVehicle(r.Context(), v); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, v)
}

func (h *Handler) GetVehicle(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.Require(w, r)
	if !ok {
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Vehicle ID required")
		return
	}

	v, err := h.repo.GetVehicleByID(r.Context(), id, sess.UserID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	if v == nil {
		response.Error(w, http.StatusNotFound, "Vehicle not found")
		return
	}

	v.ApplyDerived(time.Now())
	response.JSON(w, http.StatusOK, v)
}

func (h *Handler) DeleteVehicle(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.Require(w, r)
	if !ok {
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Vehicle ID required")
		return
	}

	if err := h.repo.DeleteVehicle(r.Context(), id, sess.UserID); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to delete vehicle")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Vehicle deleted successfully"})
}

// ownsVehicle is the single ownership gate for the sub-resource writers below.
// They take the vehicle id straight from the URL, so without this any logged-in
// user could write mileage, PUC or insurance onto anyone else's vehicle. It
// answers with the same 404 as an unknown id, so a wrong owner cannot probe
// which ids exist. Response is already written when it returns false.
func (h *Handler) ownsVehicle(w http.ResponseWriter, r *http.Request, vehicleID, userID string) bool {
	v, err := h.repo.GetVehicleByID(r.Context(), vehicleID, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to verify vehicle")
		return false
	}
	if v == nil {
		response.Error(w, http.StatusNotFound, "Vehicle not found")
		return false
	}
	return true
}

type AddMileageRequest struct {
	Mileage    int      `json:"mileage"`
	Odometer   int      `json:"odometer"`
	Date       string   `json:"date"`
	Notes      string   `json:"notes"`
	FuelLitres *float64 `json:"fuel_litres"`
	Litres     *float64 `json:"litres"`
	FuelAmount *float64 `json:"fuel_amount"`
	Cost       *float64 `json:"cost"`
}

func (h *Handler) AddMileage(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.Require(w, r)
	if !ok {
		return
	}

	vehicleID := chi.URLParam(r, "id")
	if !h.ownsVehicle(w, r, vehicleID, sess.UserID) {
		return
	}

	var req AddMileageRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if req.Mileage <= 0 && req.Odometer > 0 {
		req.Mileage = req.Odometer
	}
	if req.Mileage <= 0 {
		response.Error(w, http.StatusBadRequest, "Valid mileage is required")
		return
	}

	if req.FuelLitres == nil && req.Litres != nil {
		req.FuelLitres = req.Litres
	}
	if req.FuelAmount == nil && req.Cost != nil {
		req.FuelAmount = req.Cost
	}

	entryDate := time.Now()
	if req.Date != "" {
		if parsed, err := parseDate(req.Date); err == nil {
			entryDate = parsed
		}
	}

	entry := &MileageEntry{
		VehicleID:  vehicleID,
		Mileage:    req.Mileage,
		Date:       entryDate,
		Notes:      req.Notes,
		FuelLitres: req.FuelLitres,
		FuelAmount: req.FuelAmount,
	}

	if err := h.repo.AddMileageEntry(r.Context(), entry); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to add mileage entry")
		return
	}

	response.JSON(w, http.StatusCreated, entry)
}

func (h *Handler) UpdateMileage(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.Require(w, r)
	if !ok {
		return
	}

	vehicleID := chi.URLParam(r, "id")
	if !h.ownsVehicle(w, r, vehicleID, sess.UserID) {
		return
	}

	mileageID := chi.URLParam(r, "mileageId")
	if mileageID == "" {
		response.Error(w, http.StatusBadRequest, "Mileage ID is required")
		return
	}

	var req AddMileageRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if req.Mileage <= 0 && req.Odometer > 0 {
		req.Mileage = req.Odometer
	}
	if req.Mileage <= 0 {
		response.Error(w, http.StatusBadRequest, "Valid mileage is required")
		return
	}

	if req.FuelLitres == nil && req.Litres != nil {
		req.FuelLitres = req.Litres
	}
	if req.FuelAmount == nil && req.Cost != nil {
		req.FuelAmount = req.Cost
	}

	entryDate := time.Now()
	if req.Date != "" {
		if parsed, err := parseDate(req.Date); err == nil {
			entryDate = parsed
		}
	}

	entry := &MileageEntry{
		ID:         mileageID,
		VehicleID:  vehicleID,
		Mileage:    req.Mileage,
		Date:       entryDate,
		Notes:      req.Notes,
		FuelLitres: req.FuelLitres,
		FuelAmount: req.FuelAmount,
	}

	if err := h.repo.UpdateMileageEntry(r.Context(), entry); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.Error(w, http.StatusNotFound, "Mileage entry not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to update mileage entry")
		return
	}

	response.JSON(w, http.StatusOK, entry)
}

func (h *Handler) DeleteMileage(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.Require(w, r)
	if !ok {
		return
	}

	vehicleID := chi.URLParam(r, "id")
	if !h.ownsVehicle(w, r, vehicleID, sess.UserID) {
		return
	}

	mileageID := chi.URLParam(r, "mileageId")
	if mileageID == "" {
		response.Error(w, http.StatusBadRequest, "Mileage ID is required")
		return
	}

	if err := h.repo.DeleteMileageEntry(r.Context(), mileageID, vehicleID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.Error(w, http.StatusNotFound, "Mileage entry not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to delete mileage entry")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Mileage entry deleted"})
}

type UpsertPUCRequest struct {
	CertificateNumber string `json:"certificate_number"`
	DocumentNumber    string `json:"document_number"`
	ExpiryDate        string `json:"expiry_date"`
	ExpiresAt         string `json:"expires_at"`
}

func (h *Handler) UpdatePUC(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.Require(w, r)
	if !ok {
		return
	}

	vehicleID := chi.URLParam(r, "id")
	if !h.ownsVehicle(w, r, vehicleID, sess.UserID) {
		return
	}

	var req UpsertPUCRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	expStr := req.ExpiryDate
	if expStr == "" {
		expStr = req.ExpiresAt
	}
	if expStr == "" {
		response.Error(w, http.StatusBadRequest, "Expiry date is required")
		return
	}

	expiryDate, err := parseDate(expStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	ref := req.CertificateNumber
	if ref == "" {
		ref = req.DocumentNumber
	}

	puc := &VehicleDocument{
		VehicleID:  vehicleID,
		Type:       DocumentPUC,
		Reference:  ref,
		ExpiryDate: expiryDate,
	}

	if err := h.repo.UpsertDocument(r.Context(), puc); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update PUC record")
		return
	}

	puc.ApplyStatus(time.Now())
	response.JSON(w, http.StatusOK, puc)
}

type UpsertInsuranceRequest struct {
	PolicyNumber   string `json:"policy_number"`
	DocumentNumber string `json:"document_number"`
	Provider       string `json:"provider"`
	ExpiryDate     string `json:"expiry_date"`
	ExpiresAt      string `json:"expires_at"`
}

func (h *Handler) UpdateInsurance(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.Require(w, r)
	if !ok {
		return
	}

	vehicleID := chi.URLParam(r, "id")
	if !h.ownsVehicle(w, r, vehicleID, sess.UserID) {
		return
	}

	var req UpsertInsuranceRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	expStr := req.ExpiryDate
	if expStr == "" {
		expStr = req.ExpiresAt
	}
	if expStr == "" {
		response.Error(w, http.StatusBadRequest, "Expiry date is required")
		return
	}

	expiryDate, err := parseDate(expStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	policy := req.PolicyNumber
	if policy == "" {
		policy = req.DocumentNumber
	}

	ins := &VehicleDocument{
		VehicleID:  vehicleID,
		Type:       DocumentInsurance,
		Reference:  policy,
		Issuer:     req.Provider,
		ExpiryDate: expiryDate,
	}

	if err := h.repo.UpsertDocument(r.Context(), ins); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update Insurance record")
		return
	}

	ins.ApplyStatus(time.Now())
	response.JSON(w, http.StatusOK, ins)
}

func parseDate(s string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02T15:04:05.000Z", s)
}
