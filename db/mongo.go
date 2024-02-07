package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoClient global variable to access MongoDB
var MongoClient *mongo.Client

// ConnectToMongo function to initialize MongoDB connection
func ConnectToMongo(uri string) {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Set global MongoClient
	MongoClient = client

	log.Println("Connected to MongoDB")
}

// GetMongoClient returns the MongoDB client
func GetMongoClient() *mongo.Client {
	return MongoClient
}
