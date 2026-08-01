package routes

import (
	"context"
	"crypto/tls"
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(MongoDb)
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	clientOpts.SetTLSConfig(tlsConfig)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Println("MongoDB connection error:", err)
		return nil
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Println("Warning: Failed to ping MongoDB Atlas:", err)
	} else {
		fmt.Println("Connected to MongoDB Atlas!")
	}

	return client
}

var Client *mongo.Client

func GetClient() *mongo.Client {
	if Client == nil {
		Client = DBinstance()
	}
	return Client
}

// OpenCollection is a function makes a connection with a collection :
func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	c := GetClient()
	if c == nil {
		return nil
	}
	return c.Database("ClusterRestaurantApp01").Collection(collectionName)
}

