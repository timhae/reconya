package systemstatus

import (
	"context"
	"reconya-ai/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SystemStatusService struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewSystemStatusService(client *mongo.Client, dbName string, collectionName string) *SystemStatusService {
	collection := client.Database(dbName).Collection(collectionName)
	return &SystemStatusService{
		client:     client,
		collection: collection,
	}
}

func (s *SystemStatusService) GetLatest() (*models.SystemStatus, error) {
	var systemStatus models.SystemStatus
	opts := options.FindOne().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	err := s.collection.FindOne(context.Background(), bson.D{}, opts).Decode(&systemStatus)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &systemStatus, nil
}

func (s *SystemStatusService) CreateOrUpdate(systemStatus *models.SystemStatus) (*models.SystemStatus, error) {
	now := time.Now()

	update := bson.M{
		"$set": bson.M{
			"updated_at": now,
			"network_id": systemStatus.NetworkID,
			"public_ip":  *systemStatus.PublicIP,
		},
		"$setOnInsert": bson.M{
			"created_at": now,
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
