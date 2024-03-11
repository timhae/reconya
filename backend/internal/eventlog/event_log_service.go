package eventlog

import (
	"context"           // Replace with the correct import path
	"reconya-ai/models" // Replace with the correct import path

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EventLogService struct
type EventLogService struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// NewEventLogService creates a new EventLogService
func NewEventLogService(client *mongo.Client, dbName string, collectionName string) *EventLogService {
	collection := client.Database(dbName).Collection(collectionName)
	return &EventLogService{
		client:     client,
		collection: collection,
	}
}

// GetAll retrieves a limited number of event logs
func (s *EventLogService) GetAll(limitSize int64) ([]models.EventLog, error) {
	var eventLogs []models.EventLog
	findOptions := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(limitSize)
	cursor, err := s.collection.Find(context.Background(), bson.D{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var eventLog models.EventLog
		if err := cursor.Decode(&eventLog); err != nil {
			return nil, err
		}
		eventLogs = append(eventLogs, eventLog)
	}
	return eventLogs, nil
}

// GetAllByDeviceId retrieves event logs for a specific device
func (s *EventLogService) GetAllByDeviceId(deviceId string, limitSize int64) ([]models.EventLog, error) {
	var eventLogs []models.EventLog
	findOptions := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(limitSize)
	cursor, err := s.collection.Find(context.Background(), bson.M{"device_id": deviceId}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var eventLog models.EventLog
		if err := cursor.Decode(&eventLog); err != nil {
			return nil, err
		}
		eventLogs = append(eventLogs, eventLog)
	}
	return eventLogs, nil
}

// CreateOne creates a single event log
func (s *EventLogService) CreateOne(eventLog *models.EventLog) error {
	_, err := s.collection.InsertOne(context.Background(), eventLog)
	return err
}

// Other methods as needed...
