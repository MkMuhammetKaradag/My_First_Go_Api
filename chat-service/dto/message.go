package dto

import (
	"time"

	"github.com/go-playground/validator/v10"
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

type GetChatMessagesInput struct {
	ChatID         primitive.ObjectID `json:"chatId" validate:"required"`
	Page           int                `json:"page" validate:"min=1"`
	Limit          int                `json:"limit" validate:"min=1"`
	ExtraPassValue int                `json:"extraPassValue" validate:"min=0"`
}
type BaseUser struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Username  string             `json:"username" bson:"username" validate:"required,min=3,max=30"`
	Email     string             `json:"email" bson:"email" validate:"required,email"`
	FirstName string             `json:"firstName" bson:"firstName" validate:"required,min=3,max=50"`
}
type GetChatMessagesObject struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Sender BaseUser           `json:"sender" bson:"sender" `
	// Chat      primitive.ObjectID `json:"chat" `
	Content   string    `json:"content"  bson:"content"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

func (input *GetChatMessagesInput) Validate() error {
	validate := validator.New()
	return validate.Struct(input)
}

func (input *GetChatMessagesInput) SetDefaults() {
	if input.Page == 0 {
		input.Page = 1
	}
	if input.Limit == 0 {
		input.Limit = 10
	}
}
