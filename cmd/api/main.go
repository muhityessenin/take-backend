package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"warehouse-backend/internal/db"
	"warehouse-backend/internal/handler"
	"warehouse-backend/internal/middleware"
	"warehouse-backend/internal/repo"
	"warehouse-backend/internal/service"
)

func main() {
	r := gin.Default()

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

	itemHandler := handler.NewItemHandler(database)

	userRepo := repo.NewUserRepo(database)
	userService := service.NewUserService(userRepo)
	authHandler := handler.NewAuthHandler(userService)

	jwtService := service.NewJWTService()

	api := r.Group("/api")
	{
		// Public
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)

		// Protected
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(jwtService))
		{
			protected.GET("/items", itemHandler.GetItems)
			protected.POST("/items", itemHandler.AddItem)
			protected.PATCH("/items/:id", itemHandler.UpdateItem)

			protected.POST("/sale", itemHandler.MakeSale)
			protected.GET("/sales/today", itemHandler.GetTodaySales)
			protected.GET("/sales/top5", itemHandler.GetTop5BestSellers)
			protected.GET("/sales", itemHandler.GetSales)
		}
	}

	log.Println("🚀 Server running at :8080")
	r.Run(":8080")
}
