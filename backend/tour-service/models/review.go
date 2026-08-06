package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Review struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TourID         primitive.ObjectID `json:"tourId" bson:"tourId"`
	AuthorId       int64              `json:"authorId" bson:"authorId"`
	AuthorUsername string             `json:"authorUsername" bson:"authorUsername"`
	Rating         int                `json:"rating" bson:"rating"`
	Comment        string             `json:"comment" bson:"comment"`
	VisitDate      time.Time          `json:"visitDate" bson:"visitDate"`
	Images         []string           `json:"images" bson:"images"`
	CreatedAt      time.Time          `json:"createdAt" bson:"createdAt"`
}
