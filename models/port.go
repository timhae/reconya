package models

// Port represents the network port information
type Port struct {
	Number   string `bson:"number" json:"number"`     // Port number (e.g., "80")
	Protocol string `bson:"protocol" json:"protocol"` // Protocol (e.g., "tcp")
	State    string `bson:"state" json:"state"`       // State (e.g., "open")
	Service  string `bson:"service" json:"service"`   // Service name (e.g., "http")
}
