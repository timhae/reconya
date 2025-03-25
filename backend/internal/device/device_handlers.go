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

func (h *DeviceHandlers) GetAllDevices(w http.ResponseWriter, r *http.Request) {
	devices := []models.Device{}
	log.Printf("CIDR: %s", h.Config.NetworkCIDR)
	foundDevices, err := h.Service.FindAllForNetwork(h.Config.NetworkCIDR)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if foundDevices != nil {
		devices = foundDevices
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}
