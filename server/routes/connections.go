package routes

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DBinstance func
func DBinstance() *mongo.Client {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Note: .env file not found or failed to load, reading system env vars")
	}

	MongoDb := os.Getenv("MONGODB_URL")
	if MongoDb == "" {
		MongoDb = "mongodb://localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoDb))
	if err != nil {
		log.Println("MongoDB connection error:", err)
		return nil
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Println("Warning: Failed to ping MongoDB:", err)
	} else {
		fmt.Println("Connected to MongoDB!")
	}

	return client
}

// Client Database instance
var Client *mongo.Client = DBinstance()

// OpenCollection is a function makes a connection with a collection :
func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	if client == nil {
		return nil
	}
	var collection *mongo.Collection = client.Database("ClusterRestaurantApp01").Collection(collectionName)
	return collection
}

