// dto/chat_dto.go
package dto

import (
	"time"

	"github.com/MKMuhammetKaradag/go-microservice/shared/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateChatDto struct {
	ChatName     string               `json:"chatName" binding:"required,min=1,max=100"`
	Participants []primitive.ObjectID `json:"participants" binding:"required,min=1"`
	Admins       []primitive.ObjectID `json:"admins,omitempty"`
}

type GetChatUsersDto struct {
	ChatID primitive.ObjectID `json:"chatId"`
}

type ChatDto struct {
	ID           primitive.ObjectID   `json:"id"`
	ChatName     string               `json:"chatName"`
	Participants []primitive.ObjectID `json:"participants"`
	Admins       []primitive.ObjectID `json:"admins,omitempty"`
	CreatedAt    time.Time            `json:"createdAt"`
	UpdatedAt    time.Time            `json:"updatedAt,omitempty"`
}

type ChatWithUsers struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ChatName     string             `json:"chatName" bson:"chatName"`
	Participants []models.User      `json:"participants"`     // User objelerine değiştirildi
	Admins       []models.User      `json:"admins,omitempty"` // User objelerine değiştirildi
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}
