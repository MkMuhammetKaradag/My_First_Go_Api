package services

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/MKMuhammetKaradag/go-microservice/auth-service/database"
	"github.com/MKMuhammetKaradag/go-microservice/auth-service/dto"
	"github.com/MKMuhammetKaradag/go-microservice/auth-service/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	collection *mongo.Collection
}

func NewAuthService() *AuthService {
	return &AuthService{
		collection: database.MongoClient.Database("authDB").Collection("users"),
	}
}

func (s *AuthService) CheckExistingUser(email, username string) (bool, error) {
	filter := bson.M{"$or": []bson.M{
		{"email": email},
		{"username": username},
	}}
	count, err := s.collection.CountDocuments(context.Background(), filter)
	return count > 0, err
}

func (s *AuthService) Register(user *models.User) (*dto.UserResponse, error) {
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

// const ACTIVATION_CODE_LENGTH = 4 
func GenerateActivationCode() string {

	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(10000)

	// 4 haneye tamamlayarak string'e çevir
	return fmt.Sprintf("%04d", num)
	
}

func (s *AuthService) SignUp(user *models.User) (string, error) {
	secretKey := "your-secret-key"

	jwtService := NewJwtHelperService(secretKey)
	activationCode := GenerateActivationCode()
	payload := map[string]interface{}{
		"activationCode": activationCode,
		"user":           user,
	}
	token, err := jwtService.SignToken(payload, 1*time.Hour)
	if err != nil {
		log.Fatalf("Error signing token: %v", err)
	}

	fmt.Println("Activation Code:", activationCode)
	return token, nil
}
func ptrToInt(i int) *int {
	return &i
}
func (s *AuthService) ActivationUser(activationCode, activationToken string) (*dto.UserResponse, error) {
	secretKey := "your-secret-key"

	jwtService := NewJwtHelperService(secretKey)
	claims, err := jwtService.VerifyToken(activationToken)
	if err != nil {
		log.Fatalf("Error verifying token: %v", err)
	}

	if claims["activationCode"] != activationCode {
		return nil, errors.New("activation code mismatch")
	}

	userData := claims["user"].(map[string]interface{})
	user := &models.User{
		Username:  userData["username"].(string),
		Email:     userData["email"].(string),
		Password:  userData["password"].(string),
		FirstName: userData["firstName"].(string),
		LastName:  userData["lastName"].(string),
		Age:       ptrToInt(int(userData["age"].(float64))),
	}

	roles, ok := userData["roles"].([]interface{})
	if !ok {
		roles = []interface{}{models.USER}
	}

	var userRoles []models.UserRole
	for _, role := range roles {
		roleStr, ok := role.(string)
		if ok {
			fmt.Println("roleStr:", roleStr)
			userRoles = append(userRoles, models.UserRole(roleStr))
		}
	}

	user.Roles = userRoles
	fmt.Println("rolstrt:", roles, "user roles", userRoles)

	return s.Register(user)
}
