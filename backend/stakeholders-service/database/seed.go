package database

import (
	"log"
	"stakeholders-service/models"

	"golang.org/x/crypto/bcrypt"
)

func SeedAdmin() {
	var count int64
	DB.Model(&models.User{}).Where("role = ?", "admin").Count(&count)
	if count > 0 {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash admin password")
	}

	admin := models.User{
		Username:  "admin",
		Password:  string(hash),
		Email:     "admin@example.com",
		Role:      "admin",
		FirstName: "Admin",
		LastName:  "User",
	}

	if err := DB.Create(&admin).Error; err != nil {
		log.Printf("Admin seed failed: %v", err)
	} else {
		log.Println("Admin seeded successfully")
	}
}
