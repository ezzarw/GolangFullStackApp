package routes

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strings"
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
	} else if strings.HasPrefix(MongoDb, "mongodb+srv://") {
		cleanURI := strings.TrimSuffix(MongoDb, "/")
		parts := strings.SplitN(cleanURI, "mongodb+srv://", 2)
		if len(parts) == 2 {
			credentialsAndDomain := parts[1]
			if !strings.Contains(credentialsAndDomain, "/") {
				cleanURI += "/ClusterRestaurantApp01?retryWrites=true&w=majority&tls=true&tlsInsecure=true"
			} else if !strings.Contains(cleanURI, "retryWrites") {
				if strings.Contains(cleanURI, "?") {
					cleanURI += "&retryWrites=true&w=majority&tls=true&tlsInsecure=true"
				} else {
					cleanURI += "?retryWrites=true&w=majority&tls=true&tlsInsecure=true"
				}
			}
		}
		MongoDb = cleanURI
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(MongoDb).
		SetServerSelectionTimeout(10 * time.Second).
		SetConnectTimeout(10 * time.Second)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
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

