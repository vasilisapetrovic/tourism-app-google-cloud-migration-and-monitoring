package database

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var DB *mongo.Database

func Connect() {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Printf("Failed to create MongoDB client: %v", err)
		Client = nil
		DB = nil
		return
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Printf("Failed to ping MongoDB: %v", err)
		Client = nil
		DB = nil
		return
	}

	Client = client
	DB = client.Database("tours_db")
	log.Println("Connected to MongoDB and pinged successfully!")
}

func GetCollection(name string) *mongo.Collection {
	if DB == nil {
		log.Println("ERROR: MongoDB not connected!")
		return nil
	}
	return DB.Collection(name)
}