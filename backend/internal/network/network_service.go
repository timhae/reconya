package network

import (
	"context"           // Replace with the correct import path
	"reconya-ai/models" // Replace with the correct import path

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// NetworkService struct
// type NetworkService struct {
// 	client     *mongo.Client
// 	collection *mongo.Collection
// }

type NetworkService struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// NewNetworkService creates a new NetworkService
func NewNetworkService(client *mongo.Client, dbName string, collectionName string) *NetworkService {
	collection := client.Database(dbName).Collection(collectionName)
	return &NetworkService{
		client:     client,
		collection: collection,
	}
}

// Create a new network
func (s *NetworkService) Create(cidr string) (*models.Network, error) {
	network := &models.Network{CIDR: cidr}
	_, err := s.collection.InsertOne(context.Background(), network)
	if err != nil {
		return nil, err
	}
	return network, nil
}

// FindOrCreate finds an existing network or creates a new one
func (s *NetworkService) FindOrCreate(cidr string) (*models.Network, error) {
	var network models.Network
	err := s.collection.FindOne(context.Background(), bson.M{"cidr": cidr}).Decode(&network)
	if err == mongo.ErrNoDocuments {
		return s.Create(cidr)
	}
	if err != nil {
		return nil, err
	}
	return &network, nil
}

// FindByCIDR finds an existing network by its CIDR
func (s *NetworkService) FindByCIDR(cidr string) (*models.Network, error) {
	var network models.Network
	err := s.collection.FindOne(context.Background(), bson.M{"cidr": cidr}).Decode(&network)
	if err == mongo.ErrNoDocuments {
		return nil, nil // No network found
	}
	if err != nil {
		return nil, err // Some other error occurred
	}
	return &network, nil
}
