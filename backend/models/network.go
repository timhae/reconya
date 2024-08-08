package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Network represents a network entity
type Network struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	CIDR string             `bson:"cidr"`
}
