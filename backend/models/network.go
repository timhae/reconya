package models

import "time"

// Network represents a network entity
type Network struct {
	ID             string     `bson:"_id,omitempty" json:"id"`
	Name           string     `bson:"name" json:"name"`
	CIDR           string     `bson:"cidr" json:"cidr"`
	Description    string     `bson:"description" json:"description"`
	Status         string     `bson:"status" json:"status"` // active, inactive, scanning
	LastScannedAt  *time.Time `bson:"last_scanned_at" json:"last_scanned_at"`
	DeviceCount    int        `bson:"device_count" json:"device_count"`
	CreatedAt      time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `bson:"updated_at" json:"updated_at"`
}
