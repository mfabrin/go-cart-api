package main

import (
	"log"
	"os"

	"cart-api/internal/cart"

	"github.com/gin-gonic/gin"
)

func main() {
	ms := cart.NewMemoryStore()
	handler := cart.NewHandler(ms)

	router := gin.Default()
	api := router.Group("/api/v1")

	api.GET("/cart/:ecommerce_id/:customer_id", handler.GetCart)
	api.POST("/cart/create", handler.CreateCart)
	api.PUT("/cart/update", handler.UpdateCart)
	api.PUT("/cart/checkout", handler.CheckoutCart)
	api.DELETE("/cart/:ecommerce_id/:customer_id", handler.DeleteCart)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8081"
	}

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Unable to start the server %v", err)
	}
}
