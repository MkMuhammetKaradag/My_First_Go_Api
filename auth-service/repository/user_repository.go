package repository

import (
	"context"
	"errors"
	"time"

	"github.com/MKMuhammetKaradag/go-microservice/shared/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(collection *mongo.Collection) *UserRepository {
	return &UserRepository{collection: collection}
}

// Kullanıcı durumunu güncelleme
func (r *UserRepository) UpdateUserStatus(userID string, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kullanıcı ID'sini ObjectID'ye çevir
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("geçersiz kullanıcı ID'si")
	}

	// Durumu güncelle
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{"status": status}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return errors.New("kullanıcı bulunamadı veya durum güncellenmedi")
	}

	return nil
}

// Kullanıcı bilgilerini ID'ye göre çekme
func (r *UserRepository) FindUserByID(userID string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kullanıcı ID'sini ObjectID'ye çevir
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("geçersiz kullanıcı ID'si")
	}

	// Kullanıcıyı bul
	var user models.User
	filter := bson.M{"_id": objID}
	err = r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("kullanıcı bulunamadı")
		}
		return nil, err
	}

	return &user, nil
}

// Kullanıcıyı e-posta ile bulma
func (r *UserRepository) FindUserByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	filter := bson.M{"email": email}
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("kullanıcı bulunamadı")
		}
		return nil, err
	}

	return &user, nil
}

// Tüm kullanıcıları listeleme (örnek)
func (r *UserRepository) FindAllUsers() ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var users []models.User
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}
