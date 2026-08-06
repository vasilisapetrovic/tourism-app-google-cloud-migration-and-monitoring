package handlers

import (
	"net/http"
	"strconv"

	"stakeholders-service/database"
	"stakeholders-service/models"

	"github.com/gin-gonic/gin"
)

type LocationRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

type LocationResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	UpdatedAt string  `json:"updatedAt"`
}

// SavePosition - POST /api/position
// Cuva trenutnu lokaciju turiste
func SavePosition(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var req LocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"latitude":  req.Latitude,
			"longitude": req.Longitude,
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save position"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Position saved successfully",
		"latitude":  req.Latitude,
		"longitude": req.Longitude,
	})
}

// GetUserPosition - GET /api/position/:userId
// Vraca posljednju zabelezenu lokaciju korisnika
func GetUserPosition(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	if err := database.DB.Select("latitude", "longitude", "updated_at").
		First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, LocationResponse{
		Latitude:  user.Latitude,
		Longitude: user.Longitude,
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// GetMyPosition - GET /api/position/me
// Vraca lokaciju trenutnog korisnika
func GetMyPosition(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var user models.User
	if err := database.DB.Select("latitude", "longitude", "updated_at").
		First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, LocationResponse{
		Latitude:  user.Latitude,
		Longitude: user.Longitude,
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
