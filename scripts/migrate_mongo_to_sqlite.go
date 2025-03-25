package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"reconya-ai/db"
	"reconya-ai/internal/config"
	"reconya-ai/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Check if MongoDB URI is set
	if cfg.MongoURI == "" {
		log.Fatalf("MONGODB_URI is not set in .env file. Migration cannot continue.")
	}

	// Set SQLite path
	if cfg.SQLitePath == "" {
		dataDir := "data"
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			log.Fatalf("Failed to create data directory: %v", err)
		}
		cfg.SQLitePath = filepath.Join(dataDir, cfg.DatabaseName+".db")
	}

	// Connect to MongoDB
	log.Println("Connecting to MongoDB...")
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())

	// Connect to SQLite
	log.Println("Connecting to SQLite...")
	sqliteDB, err := db.ConnectToSQLite(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("Failed to connect to SQLite: %v", err)
	}
	defer sqliteDB.Close()

	// Initialize SQLite schema
	if err := db.InitializeSchema(sqliteDB); err != nil {
		log.Fatalf("Failed to initialize SQLite schema: %v", err)
	}

	// Create repositories
	mongoNetworkRepo := db.NewMongoNetworkRepository(mongoClient, cfg.DatabaseName, "networks")
	mongoDeviceRepo := db.NewMongoDeviceRepository(mongoClient, cfg.DatabaseName, "devices")
	mongoEventLogRepo := db.NewMongoEventLogRepository(mongoClient, cfg.DatabaseName, "event_logs")
	mongoSystemStatusRepo := db.NewMongoSystemStatusRepository(mongoClient, cfg.DatabaseName, "system_status")

	sqliteNetworkRepo := db.NewSQLiteNetworkRepository(sqliteDB)
	sqliteDeviceRepo := db.NewSQLiteDeviceRepository(sqliteDB)
	sqliteEventLogRepo := db.NewSQLiteEventLogRepository(sqliteDB)
	sqliteSystemStatusRepo := db.NewSQLiteSystemStatusRepository(sqliteDB)

	// Migrate networks
	log.Println("Migrating networks...")
	migrateNetworks(mongoClient, cfg.DatabaseName, sqliteNetworkRepo)

	// Migrate devices
	log.Println("Migrating devices...")
	migrateDevices(mongoClient, cfg.DatabaseName, sqliteDeviceRepo)

	// Migrate event logs
	log.Println("Migrating event logs...")
	migrateEventLogs(mongoClient, cfg.DatabaseName, sqliteEventLogRepo)

	// Migrate system status
	log.Println("Migrating system status...")
	migrateSystemStatus(mongoClient, cfg.DatabaseName, sqliteSystemStatusRepo)

	log.Println("Migration completed successfully!")
	log.Printf("SQLite database is available at: %s", cfg.SQLitePath)
	log.Println("To use SQLite, set DATABASE_TYPE=sqlite in your .env file")
}

func migrateNetworks(client *mongo.Client, dbName string, sqliteRepo *db.SQLiteNetworkRepository) {
	collection := client.Database(dbName).Collection("networks")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Failed to fetch networks: %v", err)
	}
	defer cursor.Close(ctx)

	count := 0
	for cursor.Next(ctx) {
		var mongoNetwork struct {
			ID   string `bson:"_id"`
			CIDR string `bson:"cidr"`
		}
		if err := cursor.Decode(&mongoNetwork); err != nil {
			log.Printf("Failed to decode network: %v", err)
			continue
		}

		network := &models.Network{
			ID:   mongoNetwork.ID,
			CIDR: mongoNetwork.CIDR,
		}

		if _, err := sqliteRepo.CreateOrUpdate(ctx, network); err != nil {
			log.Printf("Failed to create network in SQLite: %v", err)
			continue
		}
		count++
	}
	log.Printf("Migrated %d networks", count)
}

func migrateDevices(client *mongo.Client, dbName string, sqliteRepo *db.SQLiteDeviceRepository) {
	collection := client.Database(dbName).Collection("devices")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Failed to fetch devices: %v", err)
	}
	defer cursor.Close(ctx)

	count := 0
	for cursor.Next(ctx) {
		var device models.Device
		if err := cursor.Decode(&device); err != nil {
			log.Printf("Failed to decode device: %v", err)
			continue
		}

		if _, err := sqliteRepo.CreateOrUpdate(ctx, &device); err != nil {
			log.Printf("Failed to create device in SQLite: %v", err)
			continue
		}
		count++
	}
	log.Printf("Migrated %d devices", count)
}

func migrateEventLogs(client *mongo.Client, dbName string, sqliteRepo *db.SQLiteEventLogRepository) {
	collection := client.Database(dbName).Collection("event_logs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Failed to fetch event logs: %v", err)
	}
	defer cursor.Close(ctx)

	count := 0
	for cursor.Next(ctx) {
		var eventLog models.EventLog
		if err := cursor.Decode(&eventLog); err != nil {
			log.Printf("Failed to decode event log: %v", err)
			continue
		}

		if err := sqliteRepo.Create(ctx, &eventLog); err != nil {
			log.Printf("Failed to create event log in SQLite: %v", err)
			continue
		}
		count++
	}
	log.Printf("Migrated %d event logs", count)
}

func migrateSystemStatus(client *mongo.Client, dbName string, sqliteRepo *db.SQLiteSystemStatusRepository) {
	collection := client.Database(dbName).Collection("system_status")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get only the latest system status
	opts := options.FindOne().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	var systemStatus models.SystemStatus
	err := collection.FindOne(ctx, bson.M{}, opts).Decode(&systemStatus)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println("No system status found to migrate")
			return
		}
		log.Fatalf("Failed to fetch system status: %v", err)
	}

	if err := sqliteRepo.Create(ctx, &systemStatus); err != nil {
		log.Fatalf("Failed to create system status in SQLite: %v", err)
	}
	log.Println("Migrated system status successfully")
}