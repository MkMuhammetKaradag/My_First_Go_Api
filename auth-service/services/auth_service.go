package services

import (
	"context"
	"encoding/hex"
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

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

type AuthService struct {
	collection              *mongo.Collection
	passwordResetCollection *mongo.Collection
}

func NewAuthService() *AuthService {
	passwordResetCollection, _ := database.GetCollection("authDB", "passwordresets")
	return &AuthService{
		collection:              database.MongoClient.Database("authDB").Collection("users"),
		passwordResetCollection: passwordResetCollection,
	}
}

func (s *AuthService) CheckExistingUser(email, username string) (bool, error) {

	// Parametrelerin doğruluğunu kontrol et
	if email == "" && username == "" {
		return false, errors.New("en az bir parametre (email veya username) gereklidir")
	}

	// Filtreyi oluştur
	var filter bson.M
	if email != "" && username != "" {
		// Her iki parametre varsa, $or operatörü ile birleştir
		filter = bson.M{"$or": []bson.M{
			{"email": email},
			{"username": username},
		}}
	} else if email != "" {
		// Yalnızca email varsa
		filter = bson.M{"email": email}
	} else {
		// Yalnızca username varsa
		filter = bson.M{"username": username}
	}

	// Belirtilen filtreye göre belge sayısını kontrol et
	count, err := s.collection.CountDocuments(context.Background(), filter)
	if err != nil {
		return false, fmt.Errorf("veritabanı hatası: %v", err)
	}

	// Eğer 1 veya daha fazla belge varsa, kullanıcı mevcut demektir
	return count > 0, nil
}

func (s *AuthService) FindUser(email, userName string) (*models.User, error) {
	// Parametrelerin doğruluğunu kontrol et
	if email == "" && userName == "" {
		return nil, errors.New("en az bir parametre (email veya username) gereklidir")
	}

	// Filtreyi oluştur
	var filter bson.M
	if email != "" && userName != "" {
		// Her iki parametre varsa, $or operatörü ile birleştir
		filter = bson.M{"$or": []bson.M{
			{"email": email},
			{"username": userName},
		}}
	} else if email != "" {
		// Yalnızca email varsa
		filter = bson.M{"email": email}
	} else {
		// Yalnızca username varsa
		filter = bson.M{"username": userName}
	}

	// Kullanıcıyı bul
	var user models.User
	err := s.collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("Kullanıcı bulunamadı") // Kullanıcı bulunamadı
		}
		return nil, fmt.Errorf("veritabanı hatası: %v", err)
	}

	return &user, nil

}

// Şifreyi hashleme fonksiyonu
func (s *AuthService) hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// Kullanıcıyı veritabanına kaydetme fonksiyonu
func (s *AuthService) saveUserToDB(ctx context.Context, user *models.User) error {
	result, err := s.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("bu email veya kullanıcı adı zaten kullanımda")
		}
		return err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}
func (s *AuthService) Register(user *models.User) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kullanıcı zaten var mı kontrol et
	exists, err := s.CheckExistingUser(user.Email, user.Username)
	if err != nil {
		return nil, fmt.Errorf("veritabanı hatası: %v", err)
	}
	if exists {
		return nil, errors.New("bu email veya kullanıcı adı zaten kullanımda")
	}

	// Şifreyi hashle
	hashedPassword, err := s.hashPassword(user.Password)
	if err != nil {
		return nil, fmt.Errorf("şifre işlenirken hata oluştu: %v", err)
	}
	user.Password = hashedPassword

	// Tarih ve saat bilgilerini ayarla
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt

	// Kullanıcıyı veritabanına ekle
	err = s.saveUserToDB(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("kullanıcı kaydedilemedi: %v", err)
	}

	return user, nil
}

// const ACTIVATION_CODE_LENGTH = 4
func GenerateActivationCode() string {

	// 0 ile 9999 arasında bir sayı üret
	num := r.Intn(10000)

	// 4 haneye tamamlayarak string'e çevir (örn: 0034, 0923, 0001)
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
		return nil, fmt.Errorf("error verifying token: %w", err)
	}

	// Aktivasyon kodunun doğruluğunu kontrol et
	if claims["activationCode"] != activationCode {
		return nil, errors.New("activation code mismatch")
	}

	// Kullanıcı verilerini al
	userData, ok := claims["user"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid user data in token")
	}
	user := &models.User{
		Username:  userData["username"].(string),
		Email:     userData["email"].(string),
		Password:  userData["password"].(string),
		FirstName: userData["firstName"].(string),
		LastName:  userData["lastName"].(string),
		Age:       ptrToInt(int(userData["age"].(float64))),
	}
	// Roller kontrolü ve eklenmesi
	roles, ok := userData["roles"].([]interface{})
	if !ok {
		roles = []interface{}{models.USER}
	}

	var userRoles []models.UserRole
	for _, role := range roles {
		roleStr, ok := role.(string)
		if ok {

			userRoles = append(userRoles, models.UserRole(roleStr))
		}
	}

	user.Roles = userRoles

	// Kullanıcıyı kaydet (aktif hale getirme)
	return s.Register(user)
}

func (s *AuthService) SignIn(input *models.User) (*dto.UserResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var user models.User

	err := s.collection.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		// E-posta ya da şifre hatalı olduğunda aynı hatayı döndür
		return nil, errors.New("E-posta veya şifre hatalı")
	}

	// Şifreyi doğrula
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return nil, errors.New("E-posta veya şifre hatalı")
	}

	response := &dto.UserResponse{
		ID:        hex.EncodeToString(user.ID[:]),
		Username:  user.Username,
		Email:     user.Email,
		Roles:     user.Roles,
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
		// Hata günlüğü ile daha açıklayıcı bir hata mesajı
		log.Printf("Password reset token creation failed for user %s: %v", userId.Hex(), err)
		return "", fmt.Errorf("şifre sıfırlama bağlantısı oluşturulurken hata oluştu")
	}

	return resetToken, nil
}

func (s *AuthService) ForgotPassword(email string) (*string, *string, error) {

	// Kullanıcıyı bul
	user, err := s.FindUser(email, "")
	if err != nil {
		return nil, nil, err
	}

	// Şifre sıfırlama token'ını oluştur
	token, err := s.GenerateForgotPasswordLink(user.ID)
	if err != nil {
		return nil, nil, err
	}

	// Şifre sıfırlama linkini oluştur
	link := fmt.Sprintf("http://localhost:8000/resetPassword?token=%s", token)

	// Link ve kullanıcı adı döndür
	return &link, &user.Username, nil

}

func (s *AuthService) ResetPassword(input *dto.ResetPasswordDto) (*string, error) {
	var passwordReset models.PasswordReset
	filter := bson.M{
		"token": input.Token,
	}
	err := s.passwordResetCollection.FindOne(context.Background(), filter).Decode(&passwordReset)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("token  Not Found")
		}
		return nil, fmt.Errorf("error finding token: %v", err)
	}

	var user models.User
	err = s.collection.FindOne(context.Background(), bson.M{"_id": passwordReset.UserID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("User Not Found")
		}
		return nil, fmt.Errorf("error finding user: %v", err)
	}
	// Şifreyi hashle
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	if err != nil {
		return nil, errors.New("An error occurred while processing the password.")
	}
	// Kullanıcı şifresini güncelle
	update := bson.M{"$set": bson.M{"password": string(hashedPassword)}}
	_, err = s.collection.UpdateOne(context.Background(), bson.M{"_id": passwordReset.UserID}, update)
	if err != nil {
		return nil, fmt.Errorf("error updating password: %v", err)
	}
	// Şifre sıfırlama kaydını sil
	_, err = s.passwordResetCollection.DeleteOne(context.Background(), bson.M{"_id": passwordReset.ID})
	if err != nil {
		return nil, err
	}
	message := "Password successfully reset"
	return &message, nil
}

func (s *AuthService) UpdateStatus(userId, status string) error {

	objID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return errors.New("geçersiz kullanıcı ID'si")
	}

	// Durumu güncelle
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{"status": status}}

	result, err := s.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return errors.New("kullanıcı bulunamadı veya durum güncellenmedi")
	}
	return nil
}
