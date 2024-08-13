package network

import (
	"encoding/json"
	"net/http"
)

type NetworkHandlers struct {
	Service *NetworkService
}

func NewNetworkHandlers(service *NetworkService) *NetworkHandlers {
	return &NetworkHandlers{Service: service}
}

func (h *NetworkHandlers) GetNetwork(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	network, err := h.Service.FindCurrent()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if network == nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(network)
}
