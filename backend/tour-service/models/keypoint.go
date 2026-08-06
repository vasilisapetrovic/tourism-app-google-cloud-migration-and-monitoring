package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Keypoint struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TourID      primitive.ObjectID `json:"tourId" bson:"tourId"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	ImageURL    string             `json:"imageUrl" bson:"imageUrl"`
	Lat         float64            `json:"lat" bson:"lat"`
	Lng         float64            `json:"lng" bson:"lng"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
}
