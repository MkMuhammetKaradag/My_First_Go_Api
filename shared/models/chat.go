package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Chat struct {
	ID           primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	ChatName     string               `json:"chatName"  bson:"chatName" validate:"required,min=3,max=30"`
	Participants []primitive.ObjectID `json:"participants" bson:"participants" validate:"required,min=1,max=20"`
	Admins       []primitive.ObjectID `json:"admins,omitempty" bson:"admins,omitempty" `
	CreatedAt    time.Time            `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time            `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}
