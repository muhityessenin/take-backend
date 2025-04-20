package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"warehouse-backend/internal/db"
	"warehouse-backend/internal/handler"
)

func main() {
	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	database := db.Connect()
	db.AutoMigrate(database)

	h := handler.NewItemHandler(database)

	api := r.Group("/api")
	{
		api.GET("/items", h.GetItems)
		api.POST("/items", h.AddItem)
		api.POST("/sale", h.MakeSale)
		api.GET("/sales/today", h.GetTodaySales)
		api.GET("/sales/top5", h.GetTop5BestSellers)
		api.GET("/sales", h.GetSales)
		api.PATCH("/items/:id", h.UpdateItem)
	}

	log.Println("🚀 Server running at :8080")
	r.Run(":8080")
}
