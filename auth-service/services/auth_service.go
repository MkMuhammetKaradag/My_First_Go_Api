package services

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	// "github.com/MKMuhammetKaradag/go-microservice/auth-service/database"
	"github.com/MKMuhammetKaradag/go-microservice/auth-service/dto"
	"github.com/google/uuid"

	"github.com/MKMuhammetKaradag/go-microservice/shared/database"
	"github.com/MKMuhammetKaradag/go-microservice/shared/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	collection              *mongo.Collection
	passwordResetCollection *mongo.Collection
}

func NewAuthService() *AuthService {
	return &AuthService{
		collection:              database.MongoClient.Database("authDB").Collection("users"),
		passwordResetCollection: database.GetCollection("authDB", "passwordresets"),
	}
}

func (s *AuthService) CheckExistingUser(email, username string) (bool, error) {
	// filter := bson.M{"$or": []bson.M{
	// 	{"email": email},
	// 	{"username": username},
	// }}
	// count, err := s.collection.CountDocuments(context.Background(), filter)
	// return count > 0, err
	// Dinamik filtre oluştur
	var filters []bson.M

	if email != "" {
		filters = append(filters, bson.M{"email": email})
	}
	if username != "" {
		filters = append(filters, bson.M{"username": username})
	}

	// En az bir filtre varsa devam et
	if len(filters) == 0 {
		return false, errors.New("en az bir parametre (email veya username) gereklidir")
	}

	// Filtreyi $or ile birleştir veya tek filtreyi doğrudan kullan
	filter := bson.M{}
	if len(filters) > 1 {
		filter["$or"] = filters
	} else {
		filter = filters[0]
	}

	count, err := s.collection.CountDocuments(context.Background(), filter)
	return count > 0, err
}

func (s *AuthService) FindUser(email, userName string) (*models.User, error) {
	var filters []bson.M

	if email != "" {
		filters = append(filters, bson.M{"email": email})
	}
	if userName != "" {
		filters = append(filters, bson.M{"username": userName})
	}

	// En az bir filtre varsa devam et
	if len(filters) == 0 {
		return nil, errors.New("en az bir parametre (email veya userName) gereklidir")
	}

	// Filtreyi $or ile birleştir veya tek filtreyi doğrudan kullan
	filter := bson.M{}
	if len(filters) > 1 {
		filter["$or"] = filters
	} else {
		filter = filters[0]
	}

	// Kullanıcıyı bul
	var user models.User
	err := s.collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("Kulllanıcı bulunamadı") // Kullanıcı bulunamadı
		}
		return nil, err
	}

	return &user, nil

}

func (s *AuthService) Register(user *models.User) (*models.User, error) {
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

	// response := &dto.UserResponse{
	// 	ID:        hex.EncodeToString(user.ID[:]),
	// 	Username:  user.Username,
	// 	Email:     user.Email,
	// 	FirstName: user.FirstName,
	// 	Age:       *user.Age,
	// 	CreatedAt: user.CreatedAt,
	// }
	return user, nil
}

// const ACTIVATION_CODE_LENGTH = 4
func GenerateActivationCode() string {

	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(10000)

	// 4 haneye tamamlayarak string'e çevir
	return fmt.Sprintf("%04d", num)

}

func (s *AuthService) SignUp(user *models.User) (string, string, error) {
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
	return activationCode, token, nil
}
func ptrToInt(i int) *int {
	return &i
}
func (s *AuthService) ActivationUser(activationCode, activationToken string) (*models.User, error) {
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

func (s *AuthService) SignIn(input *models.User) (*dto.UserResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var user models.User

	err := s.collection.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		return nil, errors.New("Geçersiz e-posta")
	}

	// Şifreyi doğrula
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return nil, errors.New("yanlış şifre")
	}

	// Redis'e oturum kaydet
	sessionKey := "session:" + hex.EncodeToString(user.ID[:])

	userData := map[string]string{
		"id":       user.ID.Hex(),
		"email":    user.Email,
		"username": user.Username,
	}

	userDataJson, err := json.Marshal(userData)
	if err != nil {
		return nil, errors.New("Kullanıcı verisi serileştirilemedi")
	}
	err = database.RedisClient.Set(sessionKey, userDataJson, 24*time.Hour).Err()
	if err != nil {
		return nil, errors.New("oturum oluşturulurken hata oluştu")
	}

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

func (s *AuthService) GenerateForgotPasswordLink(userId primitive.ObjectID) (string, error) {
	// UUID oluştur
	resetToken := uuid.NewString()

	// Token süresini belirle (şu anki zamana +1 saat)
	expiresAt := time.Now().Add(1 * time.Hour)

	// MongoDB'ye kaydedilecek veri
	passwordReset := models.PasswordReset{
		UserID:    userId,
		Token:     resetToken,
		ExpiresAt: expiresAt,
	}

	// MongoDB'ye kaydet
	_, err := s.passwordResetCollection.InsertOne(context.Background(), passwordReset)
	if err != nil {
		return "", err
	}

	return resetToken, nil
}

func (s *AuthService) ForgotPassword(email string) (*string, *string, error) {

	user, err := s.FindUser(email, "")
	if err != nil {
		return nil, nil, err
	}
	token, err := s.GenerateForgotPasswordLink(user.ID)
	if err != nil {
		return nil, nil, err
	}

	link := fmt.Sprintf("http//localhost:8000/resetPassword?:%s", string(token))
	return &link, &user.Username, nil

}
