package models

import (
	"time"
)

// SystemStatus represents the system status entity
type SystemStatus struct {
	LocalDevice Device    `bson:"local_device"`
	Network     *Network  `bson:"network,omitempty"` // Assuming Network struct is defined
	PublicIP    *string   `bson:"public_ip,omitempty"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}
