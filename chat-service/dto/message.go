package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateMessageDto, yeni bir mesaj oluşturmak için kullanılan veri transfer objesi
type CreateMessageDto struct {
	Chat    primitive.ObjectID `json:"chat" binding:"required"`
	Content string             `json:"content" binding:"required"`
}

// MessageDto, mesaj verilerini client'a döndürmek için kullanılan veri transfer objesi
type MessageDto struct {
	ID        primitive.ObjectID `json:"id"`
	Sender    primitive.ObjectID `json:"sender"`
	Chat      primitive.ObjectID `json:"chat"`
	Content   string             `json:"content"`
	CreatedAt time.Time          `json:"createdAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt,omitempty"`
	IsDeleted bool               `json:"isDeleted,omitempty"`
	DeletedAt time.Time          `json:"deletedAt,omitempty"`
}

// UpdateMessageDto, mevcut bir mesajı güncellemek için kullanılan veri transfer objesi
type UpdateMessageDto struct {
	Content string `json:"content" binding:"required"`
}

// DeleteMessageDto, bir mesajı silmek için kullanılan veri transfer objesi
type DeleteMessageDto struct {
	Chat      primitive.ObjectID `json:"chat" binding:"required"`
	MessageID primitive.ObjectID `json:"messageId" binding:"required"`
}
