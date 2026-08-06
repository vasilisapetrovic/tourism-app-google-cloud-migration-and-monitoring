package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"
	"tour-service/database"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TourPurchaseRecord struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TouristId   int64              `json:"touristId" bson:"touristId"`
	TourId      primitive.ObjectID `json:"tourId" bson:"tourId"`
	PurchasedAt time.Time          `json:"purchasedAt" bson:"purchasedAt"`
}

func RecordPurchase(c *gin.Context) {
	var body struct {
		TouristId int64    `json:"touristId"`
		TourIds   []string `json:"tourIds"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	col := database.GetCollection("tour_purchases")
	now := time.Now()
	for _, tourIdStr := range body.TourIds {
		tourObjId, err := primitive.ObjectIDFromHex(tourIdStr)
		if err != nil {
			continue
		}
		var existing TourPurchaseRecord
		if err = col.FindOne(context.Background(), bson.M{
			"touristId": body.TouristId,
			"tourId":    tourObjId,
		}).Decode(&existing); err == nil {
			continue
		}
		col.InsertOne(context.Background(), TourPurchaseRecord{
			ID:          primitive.NewObjectID(),
			TouristId:   body.TouristId,
			TourId:      tourObjId,
			PurchasedAt: now,
		})
	}
	c.JSON(http.StatusOK, gin.H{"message": "purchases recorded"})
}

func CheckPurchase(c *gin.Context) {
	touristId, err := strconv.ParseInt(c.Query("touristId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid touristId"})
		return
	}
	tourObjId, err := primitive.ObjectIDFromHex(c.Query("tourId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tourId"})
		return
	}

	col := database.GetCollection("tour_purchases")
	var record TourPurchaseRecord
	err = col.FindOne(context.Background(), bson.M{
		"touristId": touristId,
		"tourId":    tourObjId,
	}).Decode(&record)

	purchased := (err == nil || err != mongo.ErrNoDocuments) && err == nil
	c.JSON(http.StatusOK, gin.H{"purchased": purchased})
}

func IsTourPurchasedByTourist(touristId int64, tourObjId primitive.ObjectID) bool {
	col := database.GetCollection("tour_purchases")
	var record TourPurchaseRecord
	err := col.FindOne(context.Background(), bson.M{
		"touristId": touristId,
		"tourId":    tourObjId,
	}).Decode(&record)
	return err == nil
}
