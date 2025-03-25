package models

// Network represents a network entity
type Network struct {
	ID   string `bson:"_id,omitempty" json:"id"`
	CIDR string `bson:"cidr" json:"cidr"`
}
