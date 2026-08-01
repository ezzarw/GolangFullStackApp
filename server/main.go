package main

import (
	"os"

	"server/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	router := gin.New()
	router.Use(gin.Logger())

	router.Use(cors.Default())

	// Ensure local uploads directory exists
	_ = os.MkdirAll("./uploads", os.ModePerm)

	// Serve uploaded files statically
	router.Static("/uploads", "./uploads")

	// UPLOAD MEDIA
	router.POST("/upload", routes.UploadMedia)

	// these are the endpoints
	//CREATE
	router.POST("/order/create", routes.AddOrder)
	//READ
	router.GET("/waiter/:waiter", routes.GetOrdersByWaiter)
	router.GET("/orders", routes.GetOrders)
	router.GET("/order/:id/", routes.GetOrderById)
	//UPDATE
	router.PUT("/waiter/update/:id", routes.UpdateWaiter)
	router.PUT("/order/update/:id", routes.UpdateOrder)
	//DELETE
	router.DELETE("/order/delete/:id", routes.DeleteOrder)

	// this runs the server and allows it to listen to requests
	router.Run(":" + port)
}
