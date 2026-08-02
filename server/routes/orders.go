package routes

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
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

func getOrderCollection() *mongo.Collection {
	return OpenCollection(Client, "orders")
}

// upload media file to Supabase Storage with local disk fallback
func UploadMedia(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	uniqueFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseSecretKey := os.Getenv("SUPABASE_SECRET_KEY")
	supabaseBucket := os.Getenv("SUPABASE_BUCKET")

	// If Supabase environment variables are provided, try Supabase first
	if supabaseUrl != "" && supabaseSecretKey != "" && supabaseBucket != "" {
		fileBytes, err := io.ReadAll(file)
		if err == nil {
			targetURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseUrl, supabaseBucket, uniqueFileName)
			req, reqErr := http.NewRequest("POST", targetURL, bytes.NewReader(fileBytes))
			if reqErr == nil {
				req.Header.Set("Authorization", "Bearer "+supabaseSecretKey)
				req.Header.Set("apiKey", supabaseSecretKey)
				contentType := header.Header.Get("Content-Type")
				if contentType == "" {
					contentType = "application/octet-stream"
				}
				req.Header.Set("Content-Type", contentType)

				client := &http.Client{Timeout: 30 * time.Second}
				resp, doErr := client.Do(req)
				if doErr == nil {
					defer resp.Body.Close()
					if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
						publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseUrl, supabaseBucket, uniqueFileName)
						c.JSON(http.StatusOK, gin.H{
							"message":  "File uploaded successfully to Supabase",
							"url":      publicURL,
							"filename": uniqueFileName,
						})
						return
					}
				}
			}
		}
	}

	// Fallback to local server disk storage
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat direktori upload"})
		return
	}

	savePath := filepath.Join(uploadDir, uniqueFileName)
	out, err := os.Create(savePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file ke disk"})
		return
	}
	defer out.Close()

	// Reset read pointer for local file copy
	if _, err := file.Seek(0, 0); err != nil {
		log.Println("Seek error:", err)
	}

	if _, err := io.Copy(out, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menulis file"})
		return
	}

	scheme := "https"
	if c.Request.TLS == nil && c.Request.Header.Get("X-Forwarded-Proto") == "http" {
		scheme = "http"
	}
	localURL := fmt.Sprintf("%s://%s/uploads/%s", scheme, c.Request.Host, uniqueFileName)

	c.JSON(http.StatusOK, gin.H{
		"message":  "File uploaded successfully to local storage",
		"url":      localURL,
		"filename": uniqueFileName,
	})
}

// add an order
func AddOrder(c *gin.Context) {
	col := getOrderCollection()
	if col == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: Gagal terhubung ke MongoDB Atlas"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

	result, insertErr := col.InsertOne(ctx, order)
	if insertErr != nil {
		msg := fmt.Sprintf("order item was not created: %v", insertErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, result)
}

// get all orders and return all the orders within the collection
func GetOrders(c *gin.Context) {
	col := getOrderCollection()
	if col == nil {
		c.JSON(http.StatusOK, []bson.M{})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var orders []bson.M = []bson.M{}
	cursor, err := col.Find(ctx, bson.M{})

	if err != nil {
		c.JSON(http.StatusOK, []bson.M{})
		return
	}
	if err = cursor.All(ctx, &orders); err != nil {
		c.JSON(http.StatusOK, []bson.M{})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// get all orders by the waiter's name
func GetOrdersByWaiter(c *gin.Context) {
	col := getOrderCollection()
	if col == nil {
		c.JSON(http.StatusOK, []bson.M{})
		return
	}
	waiter := c.Params.ByName("waiter")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var orders []bson.M

	cursor, err := col.Find(ctx, bson.M{"server": waiter})
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
	col := getOrderCollection()
	if col == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	orderID := c.Params.ByName("id")
	docID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var order bson.M

	if err := col.FindOne(ctx, bson.M{"_id": docID}).Decode(&order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

// update a waiter's name for an order
func UpdateWaiter(c *gin.Context) {
	col := getOrderCollection()
	if col == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	orderID := c.Params.ByName("id")
	docID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	type Waiter struct {
		Server *string `json:"server"`
	}

	var waiter Waiter

	if err := c.BindJSON(&waiter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := col.UpdateOne(ctx, bson.M{"_id": docID},
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
	col := getOrderCollection()
	if col == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	orderID := c.Params.ByName("id")
	docID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

	result, err := col.ReplaceOne(
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
	col := getOrderCollection()
	if col == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	orderID := c.Params.ByName("id")
	docID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := col.DeleteOne(ctx, bson.M{"_id": docID})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result.DeletedCount)
}
