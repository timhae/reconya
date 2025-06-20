package device

import (
	"encoding/json"
	"log"
	"net/http"
	"reconya-ai/internal/config"
	"reconya-ai/models"
)

type DeviceHandlers struct {
	Service *DeviceService
	Config  *config.Config
}

func NewDeviceHandlers(service *DeviceService, cfg *config.Config) *DeviceHandlers {
	return &DeviceHandlers{Service: service, Config: cfg}
}

func (h *DeviceHandlers) CreateDevice(w http.ResponseWriter, r *http.Request) {
	var createDeviceDto models.Device
	if err := json.NewDecoder(r.Body).Decode(&createDeviceDto); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	deviceEntity := models.Device{}

	createdDevice, err := h.Service.CreateOrUpdate(&deviceEntity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdDevice)
}

func (h *DeviceHandlers) UpdateDevice(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Path[len("/devices/"):]
	if deviceID == "" {
		http.Error(w, "device ID is required", http.StatusBadRequest)
		return
	}

	var updateData struct {
		Name    *string `json:"name,omitempty"`
		Comment *string `json:"comment,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedDevice, err := h.Service.UpdateDevice(deviceID, updateData.Name, updateData.Comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedDevice)
}

func (h *DeviceHandlers) GetAllDevices(w http.ResponseWriter, r *http.Request) {
	devices := []models.Device{}
	log.Printf("CIDR: %s", h.Config.NetworkCIDR)
	// Only return devices that have been actually discovered online
	foundDevices, err := h.Service.FindOnlineDevicesForNetwork(h.Config.NetworkCIDR)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if foundDevices != nil {
		devices = foundDevices
	}

	log.Printf("Returning %d active devices (online/idle)", len(devices))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}
