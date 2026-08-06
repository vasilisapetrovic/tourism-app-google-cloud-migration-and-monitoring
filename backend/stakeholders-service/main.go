package main

import (
	"os"
	"stakeholders-service/database"
	"stakeholders-service/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	database.Connect()
	database.SeedAdmin()
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	r.Static("/uploads", "./uploads")

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.POST("/api/auth/register", handlers.Register)
	r.POST("/api/auth/login", handlers.Login)

	profile := r.Group("/api/profile")
	profile.Use(handlers.AuthMiddleware())
	{
		profile.GET("", handlers.GetProfile)
		profile.PUT("", handlers.UpdateProfile)
		profile.GET("/api/users", handlers.AuthMiddleware(), handlers.GetPublicUsers)
		profile.GET("/api/users/:id", handlers.AuthMiddleware(), handlers.GetUserByID)
		profile.POST("/picture", handlers.UploadProfilePicture)
	}

	admin := r.Group("/api/user_management")
	admin.Use(handlers.AuthMiddleware())
	{
		admin.GET("/users", handlers.GetAllUsers)
		admin.PATCH("/users/:id/block", handlers.BlockUser)
	}

	position := r.Group("/api/position")
	position.Use(handlers.AuthMiddleware())
	{
		position.POST("", handlers.SavePosition)
		position.GET("/me", handlers.GetMyPosition)
	}

	r.GET("/api/position/:userId", handlers.GetUserPosition)
	r.GET("/internal/users/:id/blocked", handlers.GetUserBlockedStatus)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	r.Run(":" + port)
}