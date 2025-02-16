package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MKMuhammetKaradag/go-microservice/chat-service/dto"
	"github.com/MKMuhammetKaradag/go-microservice/shared/database"
	"github.com/MKMuhammetKaradag/go-microservice/shared/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ChatService struct {
	userCollection    *mongo.Collection
	chatCollection    *mongo.Collection
	messageCollection *mongo.Collection
}

func NewChatService() *ChatService {
	return &ChatService{
		userCollection:    database.GetCollection("chatDB", "users"),
		chatCollection:    database.GetCollection("chatDB", "chats"),
		messageCollection: database.GetCollection("chatDB", "messages"),
	}
}

func (s *ChatService) CreateChat(input *models.Chat) (*dto.ChatDto, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	input.CreatedAt = now
	input.UpdatedAt = now
	result, err := s.chatCollection.InsertOne(ctx, input)
	if err != nil {

		return nil, errors.New(err.Error())
	}
	input.ID = result.InsertedID.(primitive.ObjectID)

	return (*dto.ChatDto)(input), nil
}
func (s *ChatService) SendMessage(input *models.Message) (*dto.MessageDto, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	input.CreatedAt = now
	input.UpdatedAt = now
	result, err := s.messageCollection.InsertOne(ctx, input)
	if err != nil {
		fmt.Println(err.Error())
		return nil, errors.New(err.Error())
	}

	input.ID = result.InsertedID.(primitive.ObjectID)

	return (*dto.MessageDto)(input), nil
}

func (s *ChatService) GetChatWithUsersAggregation(chatID primitive.ObjectID) (*dto.ChatWithUsers, error) {
	chatCollection := s.chatCollection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{

		bson.D{{Key: "$match", Value: bson.M{"_id": chatID}}},


		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         "users",
			"localField":   "participants",
			"foreignField": "_id",
			"as":           "participantDetails",
		}}},


		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         "users",
			"localField":   "admins",
			"foreignField": "_id",
			"as":           "adminDetails",
		}}},


		bson.D{{Key: "$project", Value: bson.M{
			"_id":                1,
			"chatName":           1,
			"participantDetails": 1,
			"adminDetails":       1,
			"createdAt":          1,
			"updatedAt":          1,
		}}},
	}

	cursor, err := chatCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("chat not found")
	}


	chatData := results[0]

	participantDetails, _ := chatData["participantDetails"].(primitive.A)
	adminDetails, _ := chatData["adminDetails"].(primitive.A)

	var participants []models.User
	var admins []models.User

	// fmt.Println(participantDetails)
	// fmt.Println(adminDetails)


	for _, p := range participantDetails {
		participantMap, ok := p.(bson.M)
		if ok {
			var user models.User

			bytes, _ := bson.Marshal(participantMap)
			bson.Unmarshal(bytes, &user)
			participants = append(participants, user)
		}
	}


	for _, a := range adminDetails {
		adminMap, ok := a.(bson.M)
		if ok {
			var user models.User
			bytes, _ := bson.Marshal(adminMap)
			bson.Unmarshal(bytes, &user)
			admins = append(admins, user)
		}
	}

	chat := &dto.ChatWithUsers{
		ID:           chatData["_id"].(primitive.ObjectID),
		ChatName:     chatData["chatName"].(string),
		Participants: participants,
		Admins:       admins,
		CreatedAt:    chatData["createdAt"].(primitive.DateTime).Time(),
	}

	if updatedAt, ok := chatData["updatedAt"].(primitive.DateTime); ok {
		chat.UpdatedAt = updatedAt.Time()
	}

	return chat, nil
}
