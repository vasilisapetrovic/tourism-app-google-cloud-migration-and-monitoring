package handlers

import (
	"context"
	"net/http"
	"time"
	"tour-service/database"
	"tour-service/models"
	"tour-service/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const keypointProximityMeters = 50.0

func StartExecution(c *gin.Context) {
	tourObjId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tour ID"})
		return
	}

	tourCollection := database.GetCollection("tours")
	var tour models.Tour
	if err := tourCollection.FindOne(context.Background(), bson.M{"_id": tourObjId}).Decode(&tour); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tour not found"})
		return
	}
	if tour.Status != "published" && tour.Status != "archived" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tour must be published or archived to start"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	if extractRoleFromToken(authHeader) != "tourist" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only tourists can start tour executions"})
		return
	}
	touristId := extractUserIdFromToken(authHeader)
	touristUsername := extractUsernameFromToken(authHeader)

	if !IsTourPurchasedByTourist(touristId, tourObjId) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You must purchase this tour before starting"})
		return
	}

	execCollection := database.GetCollection("executions")

	now := time.Now()
	execCollection.UpdateMany(context.Background(), bson.M{
		"touristId": touristId,
		"tourId":    tourObjId,
		"status":    "active",
	}, bson.M{
		"$set": bson.M{"status": "abandoned", "endedAt": now, "lastActivity": now},
	})

	var body struct {
		StartLat float64 `json:"startLat"`
		StartLng float64 `json:"startLng"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	execution := models.TourExecution{
		ID:                 primitive.NewObjectID(),
		TourID:             tourObjId,
		TouristId:          touristId,
		TouristUsername:    touristUsername,
		Status:             "active",
		StartedAt:          now,
		LastActivity:       now,
		CurrentLat:         body.StartLat,
		CurrentLng:         body.StartLng,
		CompletedKeypoints: []models.CompletedKeypoint{},
	}

	if _, err := execCollection.InsertOne(context.Background(), execution); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start execution"})
		return
	}

	c.JSON(http.StatusCreated, execution)
}

func UpdatePosition(c *gin.Context) {
	execObjId, err := primitive.ObjectIDFromHex(c.Param("execId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution ID"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	touristId := extractUserIdFromToken(authHeader)

	execCollection := database.GetCollection("executions")
	var execution models.TourExecution
	if err := execCollection.FindOne(context.Background(), bson.M{"_id": execObjId}).Decode(&execution); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Execution not found"})
		return
	}
	if execution.TouristId != touristId {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not your execution"})
		return
	}
	if execution.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Execution is not active"})
		return
	}

	var body struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	kpCollection := database.GetCollection("keypoints")
	cursor, err := kpCollection.Find(context.Background(), bson.M{"tourId": execution.TourID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch keypoints"})
		return
	}
	defer cursor.Close(context.Background())

	var keypoints []models.Keypoint
	cursor.All(context.Background(), &keypoints)

	completedIds := map[primitive.ObjectID]bool{}
	for _, ck := range execution.CompletedKeypoints {
		completedIds[ck.KeypointID] = true
	}

	now := time.Now()
	var newlyCompleted []models.CompletedKeypoint
	for _, kp := range keypoints {
		if completedIds[kp.ID] {
			continue
		}
		dist := utils.HaversineDistance(body.Lat, body.Lng, kp.Lat, kp.Lng)
		if dist <= keypointProximityMeters {
			newlyCompleted = append(newlyCompleted, models.CompletedKeypoint{
				KeypointID:  kp.ID,
				Name:        kp.Name,
				CompletedAt: now,
			})
		}
	}

	update := bson.M{
		"$set": bson.M{
			"currentLat":   body.Lat,
			"currentLng":   body.Lng,
			"lastActivity": now,
		},
	}
	if len(newlyCompleted) > 0 {
		update["$push"] = bson.M{
			"completedKeypoints": bson.M{"$each": newlyCompleted},
		}
	}

	execCollection.UpdateOne(context.Background(), bson.M{"_id": execObjId}, update)
	execCollection.FindOne(context.Background(), bson.M{"_id": execObjId}).Decode(&execution)

	allCompleted := len(execution.CompletedKeypoints) == len(keypoints) && len(keypoints) > 0
	if allCompleted {
		endTime := now
		execCollection.UpdateOne(context.Background(), bson.M{"_id": execObjId}, bson.M{
			"$set": bson.M{"status": "completed", "endedAt": endTime},
		})
		execution.Status = "completed"
		execution.EndedAt = &endTime
	}

	c.JSON(http.StatusOK, gin.H{
		"execution":      execution,
		"newlyCompleted": newlyCompleted,
		"allCompleted":   allCompleted,
	})
}

func AbandonExecution(c *gin.Context) {
	execObjId, err := primitive.ObjectIDFromHex(c.Param("execId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution ID"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	touristId := extractUserIdFromToken(authHeader)

	execCollection := database.GetCollection("executions")
	var execution models.TourExecution
	if err := execCollection.FindOne(context.Background(), bson.M{"_id": execObjId}).Decode(&execution); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Execution not found"})
		return
	}
	if execution.TouristId != touristId {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not your execution"})
		return
	}
	if execution.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Execution is not active"})
		return
	}

	now := time.Now()
	execCollection.UpdateOne(context.Background(), bson.M{"_id": execObjId}, bson.M{
		"$set": bson.M{"status": "abandoned", "endedAt": now, "lastActivity": now},
	})

	c.JSON(http.StatusOK, gin.H{"message": "Execution abandoned"})
}

func GetMyExecutions(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	touristId := extractUserIdFromToken(authHeader)

	execCollection := database.GetCollection("executions")
	cursor, err := execCollection.Find(context.Background(), bson.M{"touristId": touristId})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch executions"})
		return
	}
	defer cursor.Close(context.Background())

	executions := []models.TourExecution{}
	if err := cursor.All(context.Background(), &executions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse executions"})
		return
	}

	c.JSON(http.StatusOK, executions)
}

func GetExecutionById(c *gin.Context) {
	execObjId, err := primitive.ObjectIDFromHex(c.Param("execId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution ID"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	touristId := extractUserIdFromToken(authHeader)

	execCollection := database.GetCollection("executions")
	var execution models.TourExecution
	if err := execCollection.FindOne(context.Background(), bson.M{"_id": execObjId}).Decode(&execution); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Execution not found"})
		return
	}
	if execution.TouristId != touristId {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not your execution"})
		return
	}

	c.JSON(http.StatusOK, execution)
}
