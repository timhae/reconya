package models

// Network represents the network entity
type Network struct {
	CIDR string `bson:"cidr"` // MongoDB field
}
