package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SystemStatus struct {
	LocalDevice Device             `bson:"local_device"`
	NetworkID   primitive.ObjectID `bson:"network_id,omitempty"`
	PublicIP    *string            `bson:"public_ip,omitempty"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
}
