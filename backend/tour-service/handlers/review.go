package handlers

import (
	"context"
	"net/http"
	"time"
	"tour-service/database"
	"tour-service/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateReview(c *gin.Context) {
	tourObjId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tour ID"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	if extractRoleFromToken(authHeader) != "tourist" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only tourists can leave reviews"})
		return
	}
	authorId := extractUserIdFromToken(authHeader)
	authorUsername := extractUsernameFromToken(authHeader)

	var body struct {
		Rating    int      `json:"rating"`
		Comment   string   `json:"comment"`
		VisitDate string   `json:"visitDate"`
		Images    []string `json:"images"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Rating < 1 || body.Rating > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rating must be between 1 and 5"})
		return
	}
	if body.Comment == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Comment cannot be empty"})
		return
	}

	if body.VisitDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Visit date is required"})
		return
	}
	visitDate, err := time.Parse("2006-01-02", body.VisitDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid visit date format, use YYYY-MM-DD"})
		return
	}

	images := body.Images
	if images == nil {
		images = []string{}
	}

	review := models.Review{
		ID:             primitive.NewObjectID(),
		TourID:         tourObjId,
		AuthorId:       authorId,
		AuthorUsername: authorUsername,
		Rating:         body.Rating,
		Comment:        body.Comment,
		VisitDate:      visitDate,
		Images:         images,
		CreatedAt:      time.Now(),
	}

	collection := database.GetCollection("reviews")
	_, err = collection.InsertOne(context.Background(), review)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review"})
		return
	}

	c.JSON(http.StatusCreated, review)
}

func GetReviews(c *gin.Context) {
	tourObjId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tour ID"})
		return
	}

	collection := database.GetCollection("reviews")
	cursor, err := collection.Find(context.Background(), bson.M{"tourId": tourObjId})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews"})
		return
	}
	defer cursor.Close(context.Background())

	var reviews []models.Review
	if err = cursor.All(context.Background(), &reviews); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse reviews"})
		return
	}

	var avg float64
	if len(reviews) > 0 {
		total := 0
		for _, r := range reviews {
			total += r.Rating
		}
		avg = float64(total) / float64(len(reviews))
	}

	c.JSON(http.StatusOK, gin.H{
		"reviews":       reviews,
		"averageRating": avg,
		"count":         len(reviews),
	})
}
