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
func (s *ChatService) GetMyChatsWithUsersAggregation(userID string) (*dto.ChatWithUsers, error) {
	chatCollection := s.chatCollection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("geçersiz userID: %v", err)
	}

	pipeline := mongo.Pipeline{

		bson.D{{Key: "$match", Value: bson.M{"participants": userObjID}}},

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

func (s *ChatService) AddParticipants(userID string, input *dto.ChatAddParticipants) (*string, error) {

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("geçersiz userID: %v", err)
	}
	fmt.Println(input.ChatID, userID, input.Participants)
	chatFilter := bson.M{
		"_id":    input.ChatID,
		"admins": userObjID,
	}

	var existingChat struct {
		Participants []primitive.ObjectID `bson:"participants"`
	}

	err = s.chatCollection.FindOne(
		context.Background(),
		chatFilter,
	).Decode(&existingChat)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("chat bulunamadı veya admin değilsiniz")
		}
		return nil, fmt.Errorf("veritabanı hatası: %v", err)
	}

	existingParticipants := make(map[primitive.ObjectID]bool)
	for _, p := range existingChat.Participants {
		existingParticipants[p] = true
	}

	var alreadyExists []primitive.ObjectID
	var toAdd []primitive.ObjectID

	for _, newParticipant := range input.Participants {
		if existingParticipants[newParticipant] {
			alreadyExists = append(alreadyExists, newParticipant)
		} else {
			toAdd = append(toAdd, newParticipant)
		}
	}

	if len(toAdd) == 0 {
		if len(alreadyExists) > 0 {
			return nil, fmt.Errorf(
				"tüm kullanıcılar zaten ekli: %v",
				alreadyExists,
			)
		}
		return nil, fmt.Errorf("eklenecek kullanıcı yok")
	}

	update := bson.M{
		"$addToSet": bson.M{
			"participants": bson.M{"$each": toAdd},
		},
	}

	_, err = s.chatCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": input.ChatID},
		update,
	)

	if err != nil {
		return nil, fmt.Errorf("güncelleme hatası: %v", err)
	}

	successMsg := fmt.Sprintf(
		"%d kullanıcı başarıyla eklendi. Zaten ekli olanlar: %v",
		len(toAdd),
		alreadyExists,
	)

	if len(alreadyExists) > 0 {
		return &successMsg, fmt.Errorf(successMsg)
	}

	return &successMsg, nil
}
func (s *ChatService) RemoveParticipants(userID string, input *dto.ChatRemoveParticipants) (*string, error) {
	
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("geçersiz userID: %v", err)
	}


	chatFilter := bson.M{
		"_id":    input.ChatID,
		"admins": userObjID,
	}

	var existingChat struct {
		Participants []primitive.ObjectID `bson:"participants"`
	}

	err = s.chatCollection.FindOne(
		context.Background(),
		chatFilter,
	).Decode(&existingChat)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("chat bulunamadı veya admin değilsiniz")
		}
		return nil, fmt.Errorf("veritabanı hatası: %v", err)
	}


	existingParticipants := make(map[primitive.ObjectID]bool)
	for _, p := range existingChat.Participants {
		existingParticipants[p] = true
	}

	var notFound []primitive.ObjectID
	var toRemove []primitive.ObjectID

	for _, target := range input.Participants {
		if existingParticipants[target] {
			toRemove = append(toRemove, target)
		} else {
			notFound = append(notFound, target)
		}
	}


	if len(toRemove) == 0 {
		return nil, fmt.Errorf("hiçbir kullanıcı katılımcı listesinde bulunamadı")
	}
	fmt.Println(toRemove)


	if len(existingChat.Participants) == len(toRemove) {
		now := time.Now()
		update := bson.M{
			"$set": bson.M{
				"isDeleted": true,
				"deletedAt": now,
			},
			"$pull": bson.M{
				"participants": bson.M{
					"$in": toRemove,
				},
			},
		}
		_, err := s.chatCollection.UpdateOne(context.Background(), bson.M{"_id": input.ChatID}, update)
		if err != nil {
			return nil, fmt.Errorf("chat soft delete hatası: %v", err)
		}
		successMsg := fmt.Sprintf(
			"katılımcı kalmadığında chat  silindi ",
		)
		return &successMsg, nil
	}
	update := bson.M{
		"$pull": bson.M{
			"participants": bson.M{
				"$in": toRemove,
			},
		},
	}

	_, err = s.chatCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": input.ChatID},
		update,
	)

	if err != nil {
		return nil, fmt.Errorf("güncelleme hatası: %v", err)
	}

	// 6. Sonuç mesajını oluştur
	successMsg := fmt.Sprintf(
		"%d kullanıcı başarıyla silindi. Bulunamayanlar: %v",
		len(toRemove),
		notFound,
	)

	// 7. Kısmi başarı durumu
	if len(notFound) > 0 {
		return &successMsg, fmt.Errorf(successMsg)
	}

	return &successMsg, nil
}
func (s *ChatService) LeaveChat(userID, chatID string) (*string, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("geçersiz userID: %v", err)
	}
	chatObjID, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		return nil, fmt.Errorf("geçersiz chatID: %v", err)
	}

	chatFilter := bson.M{
		"_id":          chatObjID,
		"participants": userObjID,
	}

	var existingChat struct {
		Participants []primitive.ObjectID `bson:"participants"`
	}

	err = s.chatCollection.FindOne(
		context.Background(),
		chatFilter,
	).Decode(&existingChat)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("chat bulunamadı")
		}
		return nil, fmt.Errorf("veritabanı hatası: %v", err)
	}

	existingParticipants := make(map[primitive.ObjectID]bool)
	for _, p := range existingChat.Participants {
		existingParticipants[p] = true
	}

	if !existingParticipants[userObjID] {
		return nil, fmt.Errorf("kullanıcı katılımcı değil  : %v", err)

	}

	// 4. Yeni katılımcıları ekle
	// update := bson.M{
	// 	"$pull": bson.M{
	// 		"participants": bson.M{"$in": userObjID},
	// 	},
	// }

	update := bson.M{
		"$pull": bson.M{
			"participants": userObjID, // Tek bir ObjectID'yi kaldır
		},
	}
	_, err = s.chatCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": chatObjID},
		update,
	)

	if err != nil {
		return nil, fmt.Errorf("güncelleme hatası: %v", err)
	}

	successMsg := fmt.Sprintf(
		"%v kullanıcı başarıyla silindi",
		userObjID,
	)

	return &successMsg, nil
}
