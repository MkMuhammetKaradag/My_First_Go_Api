// models/user.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRole string

const (
	ADMIN   UserRole = "admin"
	COACH   UserRole = "coach"
	STUDENT UserRole = "student"
	USER    UserRole = "user"
)

type User struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Username     string             `json:"username" bson:"username" validate:"required,min=3,max=30"`
	Email        string             `json:"email" bson:"email" validate:"required,email"`
	Password     string             `json:"password" bson:"password" validate:"required,min=8"`
	FirstName    string             `json:"firstName" bson:"firstName" validate:"required,min=3,max=50"`
	LastName     string             `json:"lastName" bson:"lastName" validate:"required,min=3,max=50"`
	Age          *int               `json:"age,omitempty" bson:"age,omitempty" validate:"omitempty,min=13,max=150"`
	ProfilePhoto *string            `json:"profilePhoto,omitempty" bson:"profilePhoto,omitempty"  validate:"omitempty" `
	Roles        []UserRole         `json:"roles" bson:"roles" `
	IsDeleted    bool               `bson:"isDeleted" json:"isDeleted"`
	DeletedAt    *time.Time         `bson:"deletedAt,omitempty" json:"deletedAt,omitempty"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

func NewUser() User {
	return User{
		Roles:     []UserRole{USER}, // Default role is "user"
		IsDeleted: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
