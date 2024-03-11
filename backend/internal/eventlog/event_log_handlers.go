package eventlog

import (
	"encoding/json"
	"net/http"
	// Replace with the correct import path
)

// EventLogHandlers struct
type EventLogHandlers struct {
	Service *EventLogService // Assume EventLogService is implemented
}

// NewEventLogHandlers creates a new instance of EventLogHandlers
func NewEventLogHandlers(service *EventLogService) *EventLogHandlers {
	return &EventLogHandlers{Service: service}
}

// FindLatest handles the request to get the latest event logs
func (h *EventLogHandlers) FindLatest(w http.ResponseWriter, r *http.Request) {
	eventLogs, err := h.Service.GetAll(5) // Assuming GetAll method is implemented
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(eventLogs)
}

// FindAllByDeviceId handles the request to get event logs by device ID
func (h *EventLogHandlers) FindAllByDeviceId(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Path[len("/event-log/"):]               // Get ID from URL
	eventLogs, err := h.Service.GetAllByDeviceId(deviceID, 8) // Assuming GetAllByDeviceId method
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(eventLogs)
}
