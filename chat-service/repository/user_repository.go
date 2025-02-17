package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/MKMuhammetKaradag/go-microservice/shared/database"
	"github.com/MKMuhammetKaradag/go-microservice/shared/models"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection

func CreateUser(user *models.User) error {
	fmt.Println("kullanıcı creat oluşturma ", user)
	userCollection = database.MongoClient.Database("chatDB").Collection("users")
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := userCollection.InsertOne(context.Background(), user)
	return err
}
