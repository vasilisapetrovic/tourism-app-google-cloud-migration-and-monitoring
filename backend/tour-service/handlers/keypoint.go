package handlers

import (
	"context"
	"math"
	"net/http"
	"time"
	"tour-service/database"
	"tour-service/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func haversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func recalculateTourLength(tourObjId primitive.ObjectID) {
	ctx := context.Background()

	cursor, err := database.GetCollection("keypoints").Find(ctx, bson.M{"tourId": tourObjId})
	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	var kps []models.Keypoint
	cursor.All(ctx, &kps)

	if len(kps) < 2 {
		database.GetCollection("tours").UpdateOne(ctx,
			bson.M{"_id": tourObjId},
			bson.M{"$set": bson.M{"lengthKm": 0}},
		)
		return
	}

	var total float64
	for i := 1; i < len(kps); i++ {
		total += haversineKm(kps[i-1].Lat, kps[i-1].Lng, kps[i].Lat, kps[i].Lng)
	}

	total = math.Round(total*100) / 100

	database.GetCollection("tours").UpdateOne(ctx,
		bson.M{"_id": tourObjId},
		bson.M{"$set": bson.M{"lengthKm": total}},
	)
}

func getTourAndCheckDraft(c *gin.Context, tourId string) (*models.Tour, primitive.ObjectID, bool) {
	tourObjId, err := primitive.ObjectIDFromHex(tourId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tour ID"})
		return nil, primitive.NilObjectID, false
	}

	tourCollection := database.GetCollection("tours")
	var tour models.Tour
	err = tourCollection.FindOne(context.Background(), bson.M{"_id": tourObjId}).Decode(&tour)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tour not found"})
		return nil, primitive.NilObjectID, false
	}
	if tour.Status != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Keypoints can only be modified on draft tours"})
		return nil, primitive.NilObjectID, false
	}
	return &tour, tourObjId, true
}

func CreateKeypoint(c *gin.Context) {
	_, tourObjId, ok := getTourAndCheckDraft(c, c.Param("id"))
	if !ok {
		return
	}

	var body struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		ImageURL    string  `json:"imageUrl"`
		Lat         float64 `json:"lat"`
		Lng         float64 `json:"lng"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Name == "" || body.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name and description are required"})
		return
	}

	keypoint := models.Keypoint{
		ID:          primitive.NewObjectID(),
		TourID:      tourObjId,
		Name:        body.Name,
		Description: body.Description,
		ImageURL:    body.ImageURL,
		Lat:         body.Lat,
		Lng:         body.Lng,
		CreatedAt:   time.Now(),
	}

	collection := database.GetCollection("keypoints")
	_, err := collection.InsertOne(context.Background(), keypoint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create keypoint"})
		return
	}

	go recalculateTourLength(tourObjId)

	c.JSON(http.StatusCreated, keypoint)
}

func GetKeypoints(c *gin.Context) {
	tourObjId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tour ID"})
		return
	}

	tourCollection := database.GetCollection("tours")
	var tour models.Tour
	err = tourCollection.FindOne(context.Background(), bson.M{"_id": tourObjId}).Decode(&tour)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tour not found"})
		return
	}

	collection := database.GetCollection("keypoints")
	cursor, err := collection.Find(context.Background(), bson.M{"tourId": tourObjId})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch keypoints"})
		return
	}
	defer cursor.Close(context.Background())

	var keypoints []models.Keypoint
	if err = cursor.All(context.Background(), &keypoints); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse keypoints"})
		return
	}

	c.JSON(http.StatusOK, keypoints)
}

func UpdateKeypoint(c *gin.Context) {
	_, tourObjId, ok := getTourAndCheckDraft(c, c.Param("id"))
	if !ok {
		return
	}

	keypointObjId, err := primitive.ObjectIDFromHex(c.Param("kpId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid keypoint ID"})
		return
	}

	var body struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		ImageURL    string  `json:"imageUrl"`
		Lat         float64 `json:"lat"`
		Lng         float64 `json:"lng"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := database.GetCollection("keypoints")
	result, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": keypointObjId, "tourId": tourObjId},
		bson.M{"$set": bson.M{
			"name":        body.Name,
			"description": body.Description,
			"imageUrl":    body.ImageURL,
			"lat":         body.Lat,
			"lng":         body.Lng,
		}},
	)
	if err != nil || result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Keypoint not found"})
		return
	}

	go recalculateTourLength(tourObjId)

	var updated models.Keypoint
	collection.FindOne(context.Background(), bson.M{"_id": keypointObjId}).Decode(&updated)
	c.JSON(http.StatusOK, updated)
}

func DeleteKeypoint(c *gin.Context) {
	_, tourObjId, ok := getTourAndCheckDraft(c, c.Param("id"))
	if !ok {
		return
	}

	keypointObjId, err := primitive.ObjectIDFromHex(c.Param("kpId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid keypoint ID"})
		return
	}

	collection := database.GetCollection("keypoints")
	result, err := collection.DeleteOne(
		context.Background(),
		bson.M{"_id": keypointObjId, "tourId": tourObjId},
	)
	if err != nil || result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Keypoint not found"})
		return
	}

	go recalculateTourLength(tourObjId)

	c.JSON(http.StatusOK, gin.H{"message": "Keypoint deleted"})
}