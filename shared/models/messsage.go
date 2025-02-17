package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Sender    primitive.ObjectID `json:"sender,omitempty" bson:"sender,omitempty"`
	Chat      primitive.ObjectID `json:"chat"  bson:"chat"`
	Content   string             `json:"content" bson:"content"`
	CreatedAt time.Time          `json:"createdAt,omitempty" bson:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt,omitempty" bson:"updatedAt"`
	IsDeleted bool               `json:"isDeleted,omitempty" bson:"isDeleted,omitempty"`
	DeletedAt time.Time          `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
}
