package device

import (
	"encoding/json"
	"net/http"
	"reconya-ai/models" // Replace with the correct import path
)

// DeviceHandlers struct holds the device service
type DeviceHandlers struct {
	Service *DeviceService
}

// NewDeviceHandlers creates new device HTTP handlers
func NewDeviceHandlers(service *DeviceService) *DeviceHandlers {
	return &DeviceHandlers{Service: service}
}

// CreateDevice handles the creation of a new device
func (h *DeviceHandlers) CreateDevice(w http.ResponseWriter, r *http.Request) {
	var createDeviceDto models.Device // Define this struct based on your CreateDeviceDto
	if err := json.NewDecoder(r.Body).Decode(&createDeviceDto); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	deviceEntity := models.Device{ // Adjust this based on your DeviceEntity and DTO
		// Initialize fields from createDeviceDto
	}

	createdDevice, err := h.Service.CreateOrUpdate(&deviceEntity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdDevice)
}

// GetAllDevices handles fetching all devices
func (h *DeviceHandlers) GetAllDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := h.Service.FindAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}
