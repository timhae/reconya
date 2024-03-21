package systemstatus

import (
	"encoding/json"
	"net/http"
	// Make sure to import your models package correctly
)

// SystemStatusHandlers struct holds the system status service
type SystemStatusHandlers struct {
	Service *SystemStatusService
}

// NewSystemStatusHandlers creates new system status HTTP handlers
func NewSystemStatusHandlers(service *SystemStatusService) *SystemStatusHandlers {
	return &SystemStatusHandlers{Service: service}
}

// GetLatestSystemStatus handles fetching the latest system status
func (h *SystemStatusHandlers) GetLatestSystemStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	systemStatus, err := h.Service.GetLatest()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if systemStatus == nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(systemStatus)
}
