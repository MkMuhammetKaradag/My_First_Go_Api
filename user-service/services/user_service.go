package services

import (
	"context"
	"encoding/hex"
	"errors"
	"time"

	"github.com/MKMuhammetKaradag/go-microservice/shared/database"
	"github.com/MKMuhammetKaradag/go-microservice/shared/models"
	"github.com/MKMuhammetKaradag/go-microservice/user-service/dto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	collection *mongo.Collection
}

func NewUserService() *UserService {
	return &UserService{
		collection: database.MongoClient.Database("userDB").Collection("users"),
	}
}

func (s *UserService) CheckExistingUser(email, username string) (bool, error) {
	filter := bson.M{"$or": []bson.M{
		{"email": email},
		{"username": username},
	}}
	count, err := s.collection.CountDocuments(context.Background(), filter)
	return count > 0, err
}

func (s *UserService) Register(user *models.User) (*dto.UserResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kullanıcı zaten var mı kontrol et
	exists, err := s.CheckExistingUser(user.Email, user.Username)
	if err != nil {
		return nil, errors.New("veritabanı hatası")
	}
	if exists {
		return nil, errors.New("bu email veya kullanıcı adı zaten kullanımda")
	}

	// Şifreyi hashle
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("şifre işlenirken hata oluştu")
	}
	user.Password = string(hashedPassword)
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Kullanıcıyı veritabanına ekle
	result, err := s.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, errors.New("bu email veya kullanıcı adı zaten kullanımda")
		}
		return nil, errors.New("kullanıcı kaydedilemedi")
	}
	user.ID = result.InsertedID.(primitive.ObjectID)

	response := &dto.UserResponse{
		ID:        hex.EncodeToString(user.ID[:]),
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		Age:       *user.Age,
		CreatedAt: user.CreatedAt,
	}
	return response, nil
}
