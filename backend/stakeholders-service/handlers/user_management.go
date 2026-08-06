package handlers

import (
	"net/http"

	"stakeholders-service/database"
	"stakeholders-service/models"

	"github.com/gin-gonic/gin"
)

type UserResponse struct {
	ID             uint   `json:"id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	Role           string `json:"role"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	ProfilePicture string `json:"profilePicture"`
	Blocked        bool   `json:"blocked"`
}

func GetAllUsers(c *gin.Context) {
	role := c.MustGet("role").(string)
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin only"})
		return
	}

	var users []models.User
	if err := database.DB.Where("role != ?", "admin").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	var response []UserResponse
	for _, u := range users {
		response = append(response, UserResponse{
			ID:             u.ID,
			Username:       u.Username,
			Email:          u.Email,
			Role:           u.Role,
			FirstName:      u.FirstName,
			LastName:       u.LastName,
			ProfilePicture: u.ProfilePicture,
			Blocked:        u.Blocked,
		})
	}

	c.JSON(http.StatusOK, response)
}

func GetUserBlockedStatus(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"blocked": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"blocked": user.Blocked})
}

func BlockUser(c *gin.Context) {
	role := c.MustGet("role").(string)
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin only"})
		return
	}

	id := c.Param("id")

	var body struct {
		Blocked bool `json:"blocked"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if user.Role == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot block an admin"})
		return
	}

	database.DB.Model(&user).Update("blocked", body.Blocked)

	c.JSON(http.StatusOK, gin.H{"message": "User block status updated", "blocked": body.Blocked})
}
