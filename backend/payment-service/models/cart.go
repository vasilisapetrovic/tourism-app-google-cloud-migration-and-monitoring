package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderItem struct {
	TourId   string  `json:"tourId" bson:"tourId"`
	TourName string  `json:"tourName" bson:"tourName"`
	Price    float64 `json:"price" bson:"price"`
}

type ShoppingCart struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TouristId  int64              `json:"touristId" bson:"touristId"`
	Items      []OrderItem        `json:"items" bson:"items"`
	TotalPrice float64            `json:"totalPrice" bson:"totalPrice"`
}

type TourPurchaseToken struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TouristId   int64              `json:"touristId" bson:"touristId"`
	TourId      string             `json:"tourId" bson:"tourId"`
	TourName    string             `json:"tourName" bson:"tourName"`
	Price       float64            `json:"price" bson:"price"`
	PurchasedAt time.Time          `json:"purchasedAt" bson:"purchasedAt"`
}
