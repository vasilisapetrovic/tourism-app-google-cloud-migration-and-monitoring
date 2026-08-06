package models

import "time"

type User struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	Username       string    `json:"username" gorm:"uniqueIndex;not null"`
	Password       string    `json:"-" gorm:"not null"`
	Email          string    `json:"email" gorm:"uniqueIndex;not null"`
	Role           string    `json:"role" gorm:"not null"`
	Blocked        bool      `json:"blocked" gorm:"default:false"`
	FirstName      string    `json:"firstName"`
	LastName       string    `json:"lastName"`
	ProfilePicture string    `json:"profilePicture"`
	Biography      string    `json:"biography"`
	Motto          string    `json:"motto"`
	Latitude       float64   `json:"latitude" gorm:"default:0"`
	Longitude      float64   `json:"longitude" gorm:"default:0"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
