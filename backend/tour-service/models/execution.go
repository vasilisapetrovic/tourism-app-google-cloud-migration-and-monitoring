package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type CompletedKeypoint struct {
	KeypointID  primitive.ObjectID `json:"keypointId" bson:"keypointId"`
	Name        string             `json:"name" bson:"name"`
	CompletedAt time.Time          `json:"completedAt" bson:"completedAt"`
}

type TourExecution struct {
	ID                  primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	TourID              primitive.ObjectID   `json:"tourId" bson:"tourId"`
	TouristId           int64                `json:"touristId" bson:"touristId"`
	TouristUsername     string               `json:"touristUsername" bson:"touristUsername"`
	Status              string               `json:"status" bson:"status"`
	StartedAt           time.Time            `json:"startedAt" bson:"startedAt"`
	EndedAt             *time.Time           `json:"endedAt,omitempty" bson:"endedAt,omitempty"`
	LastActivity        time.Time            `json:"lastActivity" bson:"lastActivity"`
	CurrentLat          float64              `json:"currentLat" bson:"currentLat"`
	CurrentLng          float64              `json:"currentLng" bson:"currentLng"`
	CompletedKeypoints  []CompletedKeypoint  `json:"completedKeypoints" bson:"completedKeypoints"`
}