package routes

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"server/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var validate = validator.New()
var orderCollection *mongo.Collection = OpenCollection(Client, "orders")

// upload media file to Supabase Storage
func UploadMedia(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseSecretKey := os.Getenv("SUPABASE_SECRET_KEY")
	supabaseBucket := os.Getenv("SUPABASE_BUCKET")

	if supabaseUrl == "" || supabaseSecretKey == "" || supabaseBucket == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Konfigurasi Supabase di env belum lengkap",
		})
		return
	}

	ext := filepath.Ext(header.Filename)
	uniqueFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca konten file"})
		return
	}

	targetURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseUrl, supabaseBucket, uniqueFileName)
	req, err := http.NewRequest("POST", targetURL, bytes.NewReader(fileBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat HTTP request ke Supabase"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+supabaseSecretKey)
	req.Header.Set("apiKey", supabaseSecretKey)
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	req.Header.Set("Content-Type", contentType)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Gagal upload ke Supabase: %v", err)})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyResp, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Supabase error (%d): %s", resp.StatusCode, string(bodyResp))})
		return
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseUrl, supabaseBucket, uniqueFileName)
	c.JSON(http.StatusOK, gin.H{
		"message":  "File uploaded successfully",
		"url":      publicURL,
		"filename": uniqueFileName,
	})
}

// add an order
func AddOrder(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var order models.Order

	if err := c.BindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validationErr := validate.Struct(order)
	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
		return
	}
	order.ID = primitive.NewObjectID()

	result, insertErr := orderCollection.InsertOne(ctx, order)
	if insertErr != nil {
		msg := fmt.Sprintf("order item was not created: %v", insertErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, result)
}

// get all orders and return all the orders within the collection
func GetOrders(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var orders []bson.M
	cursor, err := orderCollection.Find(ctx, bson.M{})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err = cursor.All(ctx, &orders); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// get all orders by the waiter's name
func GetOrdersByWaiter(c *gin.Context) {
	waiter := c.Params.ByName("waiter")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var orders []bson.M

	cursor, err := orderCollection.Find(ctx, bson.M{"server": waiter})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err = cursor.All(ctx, &orders); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// get an order by it's id
func GetOrderById(c *gin.Context) {
	orderID := c.Params.ByName("id")
	docID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var order bson.M

	if err := orderCollection.FindOne(ctx, bson.M{"_id": docID}).Decode(&order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

// update a waiter's name for an order
func UpdateWaiter(c *gin.Context) {
	orderID := c.Params.ByName("id")
	docID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	type Waiter struct {
		Server *string `json:"server"`
	}

	var waiter Waiter

	if err := c.BindJSON(&waiter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := orderCollection.UpdateOne(ctx, bson.M{"_id": docID},
		bson.M{
			"$set": bson.M{"server": waiter.Server},
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result.ModifiedCount)
}

// update the order
func UpdateOrder(c *gin.Context) {
	orderID := c.Params.ByName("id")
	docID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var order models.Order
	if err := c.BindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validationErr := validate.Struct(order)
	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
		return
	}

	result, err := orderCollection.ReplaceOne(
		ctx,
		bson.M{"_id": docID},
		bson.M{
			"dish":   order.Dish,
			"price":  order.Price,
			"server": order.Server,
			"table":  order.Table,
			"image":  order.Image,
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result.ModifiedCount)
}

// delete an order given the id
func DeleteOrder(c *gin.Context) {
	orderID := c.Params.ByName("id")
	docID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	result, err := orderCollection.DeleteOne(ctx, bson.M{"_id": docID})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result.DeletedCount)
}
