package eventlog

import (
	"context" // Replace with the correct import path
	"fmt"
	"log"
	"reconya-ai/internal/device"
	"reconya-ai/models" // Replace with the correct import path
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EventLogService struct {
	client        *mongo.Client
	collection    *mongo.Collection
	DeviceService *device.DeviceService
}

func NewEventLogService(client *mongo.Client, dbName, collectionName string, deviceService *device.DeviceService) *EventLogService {
	collection := client.Database(dbName).Collection(collectionName)
	return &EventLogService{
		client:        client,
		collection:    collection,
		DeviceService: deviceService,
	}
}

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
		// Dynamically set the description based on the event type
		eventLog.Description = s.generateDescription(eventLog)
		eventLogs = append(eventLogs, eventLog)
	}
	return eventLogs, nil
}

func (s *EventLogService) generateDescription(eventLog models.EventLog) string {
	deviceInfo := "unknown device"
	if eventLog.DeviceID != nil {
		device, err := s.DeviceService.FindByID(*eventLog.DeviceID)
		if err != nil {
			log.Printf("Error fetching device information: %v", err)
		} else if device != nil && device.IPv4 != "" {
			deviceInfo = device.IPv4
		}
	}

	switch eventLog.Type {
	case models.PingSweep:
		return "Ping sweep performed"
	case models.PortScanStarted:
		return fmt.Sprintf("Port scan started for [%s]", deviceInfo)
	case models.PortScanCompleted:
		return fmt.Sprintf("Port scan completed [%s]", deviceInfo)
	case models.DeviceOnline:
		return fmt.Sprintf("Live device [%s] found", deviceInfo)
	case models.DeviceIdle:
		return fmt.Sprintf("Device [%s] became idle", deviceInfo)
	case models.DeviceOffline:
		return fmt.Sprintf("Device [%s] is now offline", deviceInfo)
	case models.LocalIPFound:
		return fmt.Sprintf("Local IPv4 address found [%s]", deviceInfo)
	case models.LocalNetworkFound:
		return "Local network found"
	case models.Warning:
		return "Warning event occurred"
	case models.Alert:
		return "Alert event occurred"
	default:
		return "Event occurred"
	}
}

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

func (s *EventLogService) CreateOne(eventLog *models.EventLog) error {
	now := time.Now()

	eventLog.CreatedAt = &now
	eventLog.UpdatedAt = &now

	_, err := s.collection.InsertOne(context.Background(), eventLog)
	return err
}
