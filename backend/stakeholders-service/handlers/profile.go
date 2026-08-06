package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"stakeholders-service/database"
	"stakeholders-service/models"

	"github.com/gin-gonic/gin"
)

func GetProfile(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             user.ID,
		"username":       user.Username,
		"email":          user.Email,
		"role":           user.Role,
		"firstName":      user.FirstName,
		"lastName":       user.LastName,
		"profilePicture": user.ProfilePicture,
		"biography":      user.Biography,
		"motto":          user.Motto,
	})
}

func UploadProfilePicture(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	file, err := c.FormFile("picture")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%d%s", userID, time.Now().Unix(), ext)
	savePath := "./uploads/" + filename

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

    baseURL := os.Getenv("BASE_URL")
    	if baseURL == "" {
    		baseURL = "http://localhost:8081"
    	}

	pictureURL := fmt.Sprintf("%s/uploads/%s", baseURL, filename)

	database.DB.Model(&models.User{}).Where("id = ?", userID).Update("profile_picture", pictureURL)

	c.JSON(http.StatusOK, gin.H{"profilePicture": pictureURL})
}

type UpdateProfileRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Biography string `json:"biography"`
	Motto     string `json:"motto"`
}

func UpdateProfile(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Biography = req.Biography
	user.Motto = req.Motto

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             user.ID,
		"username":       user.Username,
		"email":          user.Email,
		"role":           user.Role,
		"firstName":      user.FirstName,
		"lastName":       user.LastName,
		"profilePicture": user.ProfilePicture,
		"biography":      user.Biography,
		"motto":          user.Motto,
	})
}
func GetPublicUsers(c *gin.Context) {
    userID := c.MustGet("userID").(uint)
    var users []models.User
    if err := database.DB.Where("id != ? AND role != ? AND blocked = ?", userID, "admin", false).Find(&users).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
        return
    }
    var result []gin.H
    for _, u := range users {
        result = append(result, gin.H{
            "id":             u.ID,
            "username":       u.Username,
            "firstName":      u.FirstName,
            "lastName":       u.LastName,
            "profilePicture": u.ProfilePicture,
            "biography":      u.Biography,
            "role":           u.Role,
        })
    }
    c.JSON(http.StatusOK, result)
}

func GetUserByID(c *gin.Context) {
    id := c.Param("id")
    var user models.User
    if err := database.DB.First(&user, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "id":             user.ID,
        "username":       user.Username,
        "firstName":      user.FirstName,
        "lastName":       user.LastName,
        "profilePicture": user.ProfilePicture,
        "biography":      user.Biography,
        "motto":          user.Motto,
        "role":           user.Role,
    })
}