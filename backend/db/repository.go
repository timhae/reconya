package db

import (
	"context"
	"database/sql"
	"errors"
	"reconya-ai/models"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrNotFound = errors.New("record not found")
)

// Repository defines a common interface for all repositories
type Repository interface {
	Close() error
}

// NetworkRepository defines the interface for network operations
type NetworkRepository interface {
	Repository
	FindByID(ctx context.Context, id string) (*models.Network, error)
	FindByCIDR(ctx context.Context, cidr string) (*models.Network, error)
	CreateOrUpdate(ctx context.Context, network *models.Network) (*models.Network, error)
}

// DeviceRepository defines the interface for device operations
type DeviceRepository interface {
	Repository
	FindByID(ctx context.Context, id string) (*models.Device, error)
	FindByIP(ctx context.Context, ip string) (*models.Device, error)
	FindAll(ctx context.Context) ([]*models.Device, error)
	CreateOrUpdate(ctx context.Context, device *models.Device) (*models.Device, error)
	UpdateDeviceStatuses(ctx context.Context, timeout time.Duration) error
	DeleteByID(ctx context.Context, id string) error
}

// EventLogRepository defines the interface for event log operations
type EventLogRepository interface {
	Repository
	Create(ctx context.Context, eventLog *models.EventLog) error
	FindLatest(ctx context.Context, limit int) ([]*models.EventLog, error)
	FindAllByDeviceID(ctx context.Context, deviceID string) ([]*models.EventLog, error)
}

// SystemStatusRepository defines the interface for system status operations
type SystemStatusRepository interface {
	Repository
	Create(ctx context.Context, status *models.SystemStatus) error
	FindLatest(ctx context.Context) (*models.SystemStatus, error)
}

// RepositoryFactory creates repositories based on the database type
type RepositoryFactory struct {
	SQLiteDB    *sql.DB
	MongoClient *mongo.Client
	DBName      string
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(sqliteDB *sql.DB, mongoClient *mongo.Client, dbName string) *RepositoryFactory {
	return &RepositoryFactory{
		SQLiteDB:    sqliteDB,
		MongoClient: mongoClient,
		DBName:      dbName,
	}
}

// NewNetworkRepository creates a new network repository
func (f *RepositoryFactory) NewNetworkRepository() NetworkRepository {
	if f.SQLiteDB != nil {
		return NewSQLiteNetworkRepository(f.SQLiteDB)
	}
	return NewMongoNetworkRepository(f.MongoClient, f.DBName, "networks")
}

// NewDeviceRepository creates a new device repository
func (f *RepositoryFactory) NewDeviceRepository() DeviceRepository {
	if f.SQLiteDB != nil {
		return NewSQLiteDeviceRepository(f.SQLiteDB)
	}
	return NewMongoDeviceRepository(f.MongoClient, f.DBName, "devices")
}

// NewEventLogRepository creates a new event log repository
func (f *RepositoryFactory) NewEventLogRepository() EventLogRepository {
	if f.SQLiteDB != nil {
		return NewSQLiteEventLogRepository(f.SQLiteDB)
	}
	return NewMongoEventLogRepository(f.MongoClient, f.DBName, "event_logs")
}

// NewSystemStatusRepository creates a new system status repository
func (f *RepositoryFactory) NewSystemStatusRepository() SystemStatusRepository {
	if f.SQLiteDB != nil {
		return NewSQLiteSystemStatusRepository(f.SQLiteDB)
	}
	return NewMongoSystemStatusRepository(f.MongoClient, f.DBName, "system_status")
}

// GenerateID generates a unique ID for a record
func GenerateID() string {
	return uuid.New().String()
}

// ObjectIDFromString converts a string to an ObjectID
func ObjectIDFromString(id string) primitive.ObjectID {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID
	}
	return objectID
}