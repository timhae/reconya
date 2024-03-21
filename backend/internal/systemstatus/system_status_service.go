package systemstatus

import (
	"context"           // Replace with the correct import path
	"reconya-ai/models" // Replace with the correct import path
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SystemStatusService struct
type SystemStatusService struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// NewSystemStatusService creates a new SystemStatusService
func NewSystemStatusService(client *mongo.Client, dbName string, collectionName string) *SystemStatusService {
	collection := client.Database(dbName).Collection(collectionName)
	return &SystemStatusService{
		client:     client,
		collection: collection,
	}
}

// GetLatest retrieves the latest system status
func (s *SystemStatusService) GetLatest() (*models.SystemStatus, error) {
	var systemStatus models.SystemStatus
	opts := options.FindOne().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	err := s.collection.FindOne(context.Background(), bson.D{}, opts).Decode(&systemStatus)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No documents found
		}
		return nil, err
	}
	return &systemStatus, nil
}

// CreateOrUpdate creates or updates a system status
func (s *SystemStatusService) CreateOrUpdate(systemStatus *models.SystemStatus) (*models.SystemStatus, error) {
	now := time.Now()

	// Set the updated_at field to now
	update := bson.M{
		"$set": bson.M{
			"updated_at": now,
			// Set other fields that should be updated every time...
		},
		"$setOnInsert": bson.M{
			"created_at": now,
			// Set other fields that should only be set on creation...
		},
	}

	filter := bson.M{"local_device.ipv4": systemStatus.LocalDevice.IPv4}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var updatedSystemStatus models.SystemStatus
	err := s.collection.FindOneAndUpdate(context.Background(), filter, update, opts).Decode(&updatedSystemStatus)
	if err != nil {
		return nil, err
	}

	return &updatedSystemStatus, nil
}
