package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"tour-service/database"
	"tour-service/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func base64Decode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

func extractUserIdFromToken(authHeader string) int64 {
	token := strings.Replace(authHeader, "Bearer ", "", 1)
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return 0
	}
	payload, _ := base64Decode(parts[1])
	var claims map[string]interface{}
	json.Unmarshal(payload, &claims)
	id, _ := claims["id"].(float64)
	return int64(id)
}

func extractUsernameFromToken(authHeader string) string {
	token := strings.Replace(authHeader, "Bearer ", "", 1)
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return ""
	}
	payload, _ := base64Decode(parts[1])
	var claims map[string]interface{}
	json.Unmarshal(payload, &claims)
	username, _ := claims["username"].(string)
	return username
}

func extractRoleFromToken(authHeader string) string {
	token := strings.Replace(authHeader, "Bearer ", "", 1)
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return ""
	}
	payload, _ := base64Decode(parts[1])
	var claims map[string]interface{}
	json.Unmarshal(payload, &claims)
	role, _ := claims["role"].(string)
	return role
}

func CreateTour(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	authorId := extractUserIdFromToken(authHeader)
	authorUsername := extractUsernameFromToken(authHeader)

	var body struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Difficulty  string   `json:"difficulty"`
		Tags        []string `json:"tags"`
		Price       float64  `json:"price"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	price := body.Price
	if price < 0 {
		price = 0
	}

	tour := models.Tour{
		ID:             primitive.NewObjectID(),
		AuthorId:       authorId,
		AuthorUsername: authorUsername,
		Name:           body.Name,
		Description:    body.Description,
		Difficulty:     body.Difficulty,
		Tags:           body.Tags,
		Status:         "draft",
		Price:          price,
		LengthKm:       0,
		TransportTimes: []models.TransportTime{},
		CreatedAt:      time.Now(),
	}

	collection := database.GetCollection("tours")
	_, err := collection.InsertOne(context.Background(), tour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tour"})
		return
	}

	c.JSON(http.StatusCreated, tour)
}

func GetMyTours(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	authorId := extractUserIdFromToken(authHeader)

	collection := database.GetCollection("tours")
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := collection.Find(context.Background(), bson.M{"authorId": authorId}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tours"})
		return
	}
	defer cursor.Close(context.Background())

	var tours []models.Tour
	if err = cursor.All(context.Background(), &tours); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse tours"})
		return
	}

	c.JSON(http.StatusOK, tours)
}

// PublishTour
func PublishTour(c *gin.Context) {
	id := c.Param("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	authorId := extractUserIdFromToken(authHeader)

	collection := database.GetCollection("tours")
	var tour models.Tour
	err = collection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&tour)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tour not found"})
		return
	}
	if tour.AuthorId != authorId {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not your tour"})
		return
	}
	if tour.Status != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only draft tours can be published"})
		return
	}

	// Validacija - osnovni podaci
	if strings.TrimSpace(tour.Name) == "" ||
		strings.TrimSpace(tour.Description) == "" ||
		strings.TrimSpace(tour.Difficulty) == "" ||
		len(tour.Tags) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tour must have name, description, difficulty and at least one tag"})
		return
	}

	// Validacija - minimum 2 ključne tačke
	keypointCollection := database.GetCollection("keypoints")
	count, err := keypointCollection.CountDocuments(context.Background(), bson.M{"tourId": objId})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count keypoints"})
		return
	}
	if count < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tour must have at least 2 key points"})
		return
	}

	// Validacija - bar jedan transport time
	if len(tour.TransportTimes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tour must have at least one transport time (walking, bicycle or car)"})
		return
	}

	now := time.Now()
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objId},
		bson.M{"$set": bson.M{
			"status":      "published",
			"publishedAt": now,
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish tour"})
		return
	}

	tour.Status = "published"
	tour.PublishedAt = &now
	c.JSON(http.StatusOK, tour)
}

// ArchiveTour
func ArchiveTour(c *gin.Context) {
	id := c.Param("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	authorId := extractUserIdFromToken(authHeader)

	collection := database.GetCollection("tours")
	var tour models.Tour
	err = collection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&tour)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tour not found"})
		return
	}
	if tour.AuthorId != authorId {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not your tour"})
		return
	}
	if tour.Status != "published" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only published tours can be archived"})
		return
	}

	now := time.Now()
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objId},
		bson.M{"$set": bson.M{
			"status":     "archived",
			"archivedAt": now,
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to archive tour"})
		return
	}

	tour.Status = "archived"
	tour.ArchivedAt = &now
	c.JSON(http.StatusOK, tour)
}

// ReactivateTour
func ReactivateTour(c *gin.Context) {
	id := c.Param("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	authorId := extractUserIdFromToken(authHeader)

	collection := database.GetCollection("tours")
	var tour models.Tour
	err = collection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&tour)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tour not found"})
		return
	}
	if tour.AuthorId != authorId {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not your tour"})
		return
	}
	if tour.Status != "archived" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only archived tours can be reactivated"})
		return
	}

	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objId},
		bson.M{"$set": bson.M{
			"status":     "published",
			"archivedAt": nil,
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reactivate tour"})
		return
	}

	tour.Status = "published"
	tour.ArchivedAt = nil
	c.JSON(http.StatusOK, tour)
}

func AddTransportTime(c *gin.Context) {
	id := c.Param("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	authorId := extractUserIdFromToken(authHeader)

	collection := database.GetCollection("tours")
	var tour models.Tour
	err = collection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&tour)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tour not found"})
		return
	}
	if tour.AuthorId != authorId {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not your tour"})
		return
	}
	if tour.Status != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only add transport times to draft tours"})
		return
	}

	var body struct {
		TransportType string `json:"transportType"`
		Minutes       int    `json:"minutes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validTypes := map[string]bool{"walking": true, "bicycle": true, "car": true}
	if !validTypes[body.TransportType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transport type must be: walking, bicycle or car"})
		return
	}
	if body.Minutes <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Minutes must be greater than 0"})
		return
	}

	updated := false
	for i, tt := range tour.TransportTimes {
		if tt.TransportType == body.TransportType {
			tour.TransportTimes[i].Minutes = body.Minutes
			updated = true
			break
		}
	}
	if !updated {
		tour.TransportTimes = append(tour.TransportTimes, models.TransportTime{
			TransportType: body.TransportType,
			Minutes:       body.Minutes,
		})
	}

	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objId},
		bson.M{"$set": bson.M{"transportTimes": tour.TransportTimes}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transport times"})
		return
	}

	c.JSON(http.StatusOK, tour)
}

func GetPublishedTours(c *gin.Context) {
	collection := database.GetCollection("tours")
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := collection.Find(context.Background(), bson.M{"status": "published"}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tours"})
		return
	}
	defer cursor.Close(context.Background())

	var tours []models.Tour
	if err = cursor.All(context.Background(), &tours); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse tours"})
		return
	}

	c.JSON(http.StatusOK, tours)
}

func GetTourById(c *gin.Context) {
	id := c.Param("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	collection := database.GetCollection("tours")
	var tour models.Tour
	err = collection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&tour)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tour not found"})
		return
	}

	c.JSON(http.StatusOK, tour)
}

func GetTourByIdNew(c *gin.Context) {
	id := c.Param("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	collection := database.GetCollection("tours")
	var tour models.Tour
	err = collection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&tour)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tour not found"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	role := extractRoleFromToken(authHeader)
	authorId := extractUserIdFromToken(authHeader)

	if tour.AuthorId != authorId && tour.Status == "draft" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Tour not available"})
		return
	}

	_ = role
	c.JSON(http.StatusOK, tour)
}

func UpdatePrice(c *gin.Context) {
	id := c.Param("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	authorId := extractUserIdFromToken(authHeader)

	collection := database.GetCollection("tours")
	var tour models.Tour
	err = collection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&tour)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tour not found"})
		return
	}
	if tour.AuthorId != authorId {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not your tour"})
		return
	}
	if tour.Status != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price can only be changed on draft tours"})
		return
	}

	var body struct {
		Price float64 `json:"price"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Price < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price cannot be negative"})
		return
	}

	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objId},
		bson.M{"$set": bson.M{"price": body.Price}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update price"})
		return
	}

	tour.Price = body.Price
	c.JSON(http.StatusOK, tour)
}