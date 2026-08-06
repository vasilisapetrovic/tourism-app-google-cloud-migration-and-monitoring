package models

import (
    "go.mongodb.org/mongo-driver/bson/primitive"
    "time"
)

type TransportTime struct {
	TransportType string `json:"transportType" bson:"transportType"` // "walking", "bicycle", "car"
	Minutes       int    `json:"minutes" bson:"minutes"`
}

type Tour struct {
    ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    AuthorId    int64              `json:"authorId" bson:"authorId"`
    AuthorUsername string          `json:"authorUsername" bson:"authorUsername"`
    Name        string             `json:"name" bson:"name"`
    Description string             `json:"description" bson:"description"`
    Difficulty  string             `json:"difficulty" bson:"difficulty"`
    Tags        []string           `json:"tags" bson:"tags"`
    Status      string             `json:"status" bson:"status"`
    Price       float64            `json:"price" bson:"price"`
    LengthKm    float64            `json:"lengthKm" bson:"lengthKm"`
    TransportTimes []TransportTime    `json:"transportTimes" bson:"transportTimes"`
    CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
    PublishedAt *time.Time         `json:"publishedAt,omitempty" bson:"publishedAt,omitempty"`
    ArchivedAt  *time.Time         `json:"archivedAt,omitempty" bson:"archivedAt,omitempty"`
}

