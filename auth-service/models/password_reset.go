// models/password_reset.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PasswordReset struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"userId" bson:"userId" validate:"required"`
	Token     string             `json:"token" bson:"token" validate:"required"`
	ExpiresAt time.Time          `json:"expiresAt" bson:"expiresAt" validate:"required"`
	Used      bool               `json:"used" bson:"used" validate:"boolean"`
}
