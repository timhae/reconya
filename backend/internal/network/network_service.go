package network

import (
	"context"
	"reconya-ai/internal/config"
	"reconya-ai/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type NetworkService struct {
	Config     *config.Config
	Client     *mongo.Client
	Collection *mongo.Collection
}

func NewNetworkService(client *mongo.Client, dbName string, collectionName string, cfg *config.Config) *NetworkService {
	collection := client.Database(dbName).Collection(collectionName)
	return &NetworkService{
		Config:     cfg,
		Client:     client,
		Collection: collection,
	}
}
func (s *NetworkService) Create(cidr string) (*models.Network, error) {
	network := &models.Network{CIDR: cidr}
	_, err := s.Collection.InsertOne(context.Background(), network)
	if err != nil {
		return nil, err
	}
	return network, nil
}

func (s *NetworkService) FindOrCreate(cidr string) (*models.Network, error) {
	var network models.Network
	err := s.Collection.FindOne(context.Background(), bson.M{"cidr": cidr}).Decode(&network)
	if err == mongo.ErrNoDocuments {
		return s.Create(cidr)
	}
	if err != nil {
		return nil, err
	}
	return &network, nil
}

func (s *NetworkService) FindByCIDR(cidr string) (*models.Network, error) {
	var network models.Network
	err := s.Collection.FindOne(context.Background(), bson.M{"cidr": cidr}).Decode(&network)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &network, nil
}

func (s *NetworkService) FindCurrent() (*models.Network, error) {
	network, err := s.FindByCIDR(s.Config.NetworkCIDR)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return network, nil
}
