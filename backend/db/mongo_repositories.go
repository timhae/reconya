package db

import (
	"context"
	"fmt"
	"reconya-ai/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoNetworkRepository implements the NetworkRepository interface for MongoDB
type MongoNetworkRepository struct {
	client     *mongo.Client
	database   string
	collection string
}

// NewMongoNetworkRepository creates a new MongoNetworkRepository
func NewMongoNetworkRepository(client *mongo.Client, database, collection string) *MongoNetworkRepository {
	return &MongoNetworkRepository{
		client:     client,
		database:   database,
		collection: collection,
	}
}

// Close closes the MongoDB connection
func (r *MongoNetworkRepository) Close() error {
	return r.client.Disconnect(context.Background())
}

// FindByID finds a network by ID
func (r *MongoNetworkRepository) FindByID(ctx context.Context, id string) (*models.Network, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	var network models.Network
	err = r.client.Database(r.database).Collection(r.collection).
		FindOne(ctx, bson.M{"_id": objectID}).Decode(&network)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("error finding network: %w", err)
	}

	return &network, nil
}

// FindByCIDR finds a network by CIDR
func (r *MongoNetworkRepository) FindByCIDR(ctx context.Context, cidr string) (*models.Network, error) {
	var network models.Network
	err := r.client.Database(r.database).Collection(r.collection).
		FindOne(ctx, bson.M{"cidr": cidr}).Decode(&network)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("error finding network: %w", err)
	}

	return &network, nil
}

// CreateOrUpdate creates or updates a network
func (r *MongoNetworkRepository) CreateOrUpdate(ctx context.Context, network *models.Network) (*models.Network, error) {
	// If the network doesn't have an ID, create one
	if network.ID == "" {
		network.ID = primitive.NewObjectID().Hex()
	}

	opts := options.Update().SetUpsert(true)
	filter := bson.M{"_id": network.ID}
	update := bson.M{"$set": network}

	_, err := r.client.Database(r.database).Collection(r.collection).
		UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, fmt.Errorf("error upserting network: %w", err)
	}

	return network, nil
}

// MongoDeviceRepository implements the DeviceRepository interface for MongoDB
type MongoDeviceRepository struct {
	client     *mongo.Client
	database   string
	collection string
}

// NewMongoDeviceRepository creates a new MongoDeviceRepository
func NewMongoDeviceRepository(client *mongo.Client, database, collection string) *MongoDeviceRepository {
	return &MongoDeviceRepository{
		client:     client,
		database:   database,
		collection: collection,
	}
}

// Close closes the MongoDB connection
func (r *MongoDeviceRepository) Close() error {
	return r.client.Disconnect(context.Background())
}

// FindByID finds a device by ID
func (r *MongoDeviceRepository) FindByID(ctx context.Context, id string) (*models.Device, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	var device models.Device
	err = r.client.Database(r.database).Collection(r.collection).
		FindOne(ctx, bson.M{"_id": objectID}).Decode(&device)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("error finding device: %w", err)
	}

	return &device, nil
}

// FindByIP finds a device by IP address
func (r *MongoDeviceRepository) FindByIP(ctx context.Context, ip string) (*models.Device, error) {
	var device models.Device
	err := r.client.Database(r.database).Collection(r.collection).
		FindOne(ctx, bson.M{"ipv4": ip}).Decode(&device)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("error finding device: %w", err)
	}

	return &device, nil
}

// FindAll finds all devices
func (r *MongoDeviceRepository) FindAll(ctx context.Context) ([]*models.Device, error) {
	opts := options.Find().SetSort(bson.M{"updated_at": -1})
	cursor, err := r.client.Database(r.database).Collection(r.collection).
		Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("error finding devices: %w", err)
	}
	defer cursor.Close(ctx)

	var devices []*models.Device
	if err = cursor.All(ctx, &devices); err != nil {
		return nil, fmt.Errorf("error decoding devices: %w", err)
	}

	return devices, nil
}

// CreateOrUpdate creates or updates a device
func (r *MongoDeviceRepository) CreateOrUpdate(ctx context.Context, device *models.Device) (*models.Device, error) {
	now := time.Now()
	device.UpdatedAt = now

	// If the device doesn't have an ID, create one
	if device.ID == "" {
		device.ID = primitive.NewObjectID().Hex()
		device.CreatedAt = now
	}

	opts := options.Update().SetUpsert(true)
	filter := bson.M{"_id": device.ID}
	update := bson.M{"$set": device}

	_, err := r.client.Database(r.database).Collection(r.collection).
		UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, fmt.Errorf("error upserting device: %w", err)
	}

	return device, nil
}

// UpdateDeviceStatuses updates device statuses based on last seen time
func (r *MongoDeviceRepository) UpdateDeviceStatuses(ctx context.Context, timeout time.Duration) error {
	now := time.Now()
	offlineThreshold := now.Add(-timeout)

	// Update devices to offline if they haven't been seen for longer than the threshold
	filter := bson.M{
		"status":              bson.M{"$in": []string{string(models.DeviceStatusOnline), string(models.DeviceStatusIdle)}},
		"last_seen_online_at": bson.M{"$lt": offlineThreshold},
	}
	update := bson.M{
		"$set": bson.M{
			"status":     models.DeviceStatusOffline,
			"updated_at": now,
		},
	}

	_, err := r.client.Database(r.database).Collection(r.collection).
		UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("error updating device statuses to offline: %w", err)
	}

	// Update online devices to idle if they've been online for more than half the timeout
	idleThreshold := now.Add(-timeout / 2)
	filter = bson.M{
		"status":              models.DeviceStatusOnline,
		"last_seen_online_at": bson.M{"$lt": idleThreshold},
	}
	update = bson.M{
		"$set": bson.M{
			"status":     models.DeviceStatusIdle,
			"updated_at": now,
		},
	}

	_, err = r.client.Database(r.database).Collection(r.collection).
		UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("error updating device statuses to idle: %w", err)
	}

	return nil
}

// DeleteByID deletes a device by ID
func (r *MongoDeviceRepository) DeleteByID(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}

	result, err := r.client.Database(r.database).Collection(r.collection).
		DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("error deleting device: %w", err)
	}

	if result.DeletedCount == 0 {
		return ErrNotFound
	}

	return nil
}

// MongoEventLogRepository implements the EventLogRepository interface for MongoDB
type MongoEventLogRepository struct {
	client     *mongo.Client
	database   string
	collection string
}

// NewMongoEventLogRepository creates a new MongoEventLogRepository
func NewMongoEventLogRepository(client *mongo.Client, database, collection string) *MongoEventLogRepository {
	return &MongoEventLogRepository{
		client:     client,
		database:   database,
		collection: collection,
	}
}

// Close closes the MongoDB connection
func (r *MongoEventLogRepository) Close() error {
	return r.client.Disconnect(context.Background())
}

// Create creates a new event log
func (r *MongoEventLogRepository) Create(ctx context.Context, eventLog *models.EventLog) error {
	now := time.Now()
	if eventLog.CreatedAt == nil {
		eventLog.CreatedAt = &now
	}
	if eventLog.UpdatedAt == nil {
		eventLog.UpdatedAt = &now
	}

	_, err := r.client.Database(r.database).Collection(r.collection).
		InsertOne(ctx, eventLog)
	if err != nil {
		return fmt.Errorf("error creating event log: %w", err)
	}

	return nil
}

// FindLatest finds the latest event logs
func (r *MongoEventLogRepository) FindLatest(ctx context.Context, limit int) ([]*models.EventLog, error) {
	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetLimit(int64(limit))

	cursor, err := r.client.Database(r.database).Collection(r.collection).
		Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("error finding event logs: %w", err)
	}
	defer cursor.Close(ctx)

	var logs []*models.EventLog
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("error decoding event logs: %w", err)
	}

	return logs, nil
}

// FindAllByDeviceID finds all event logs for a device
func (r *MongoEventLogRepository) FindAllByDeviceID(ctx context.Context, deviceID string) ([]*models.EventLog, error) {
	opts := options.Find().SetSort(bson.M{"created_at": -1})
	cursor, err := r.client.Database(r.database).Collection(r.collection).
		Find(ctx, bson.M{"device_id": deviceID}, opts)
	if err != nil {
		return nil, fmt.Errorf("error finding device event logs: %w", err)
	}
	defer cursor.Close(ctx)

	var logs []*models.EventLog
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("error decoding event logs: %w", err)
	}

	return logs, nil
}

// MongoSystemStatusRepository implements the SystemStatusRepository interface for MongoDB
type MongoSystemStatusRepository struct {
	client     *mongo.Client
	database   string
	collection string
}

// NewMongoSystemStatusRepository creates a new MongoSystemStatusRepository
func NewMongoSystemStatusRepository(client *mongo.Client, database, collection string) *MongoSystemStatusRepository {
	return &MongoSystemStatusRepository{
		client:     client,
		database:   database,
		collection: collection,
	}
}

// Close closes the MongoDB connection
func (r *MongoSystemStatusRepository) Close() error {
	return r.client.Disconnect(context.Background())
}

// Create creates a new system status
func (r *MongoSystemStatusRepository) Create(ctx context.Context, status *models.SystemStatus) error {
	now := time.Now()
	status.CreatedAt = now
	status.UpdatedAt = now

	_, err := r.client.Database(r.database).Collection(r.collection).
		InsertOne(ctx, status)
	if err != nil {
		return fmt.Errorf("error creating system status: %w", err)
	}

	return nil
}

// FindLatest finds the latest system status
func (r *MongoSystemStatusRepository) FindLatest(ctx context.Context) (*models.SystemStatus, error) {
	opts := options.FindOne().SetSort(bson.M{"created_at": -1})
	var status models.SystemStatus
	err := r.client.Database(r.database).Collection(r.collection).
		FindOne(ctx, bson.M{}, opts).Decode(&status)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("error finding system status: %w", err)
	}

	return &status, nil
}
