package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/MKMuhammetKaradag/go-microservice/shared/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ChatRepository struct {
	collection *mongo.Collection
}

func NewChatRepository(collection *mongo.Collection) *ChatRepository {
	dbcollection, _ := database.GetCollection("chatDB", "chats")
	return &ChatRepository{collection: dbcollection}
}
func (r *ChatRepository) IsUserInChat(chatID, userID string) (bool, error) {

	chatObjID, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		return false, fmt.Errorf("geçersiz chatID: %v", err)
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, fmt.Errorf("geçersiz userID: %v", err)
	}

	filter := bson.M{
		"_id":          chatObjID,
		"participants": userObjID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
