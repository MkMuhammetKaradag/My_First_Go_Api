package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

func ConnectMongoDB(mongoURI string) {
	// Bağlantı URI'si
	// mongoURI := "mongodb://localhost:27017"

	// MongoDB Bağlantı Seçenekleri
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Bağlantıyı başlat
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatalf("MongoDB client oluşturulamadı: %v", err)
	}

	// Bağlantıyı aç
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("MongoDB'ye bağlanılamadı: %v", err)
	}

	// Bağlantıyı test et
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("MongoDB ping başarısız: %v", err)
	}

	fmt.Println("MongoDB bağlantısı başarılı!")
	MongoClient = client

	createUserCollectionWithSchema()
	createPasswordResetCollectionWithSchema()
	fmt.Println("Koleksiyonlar ve şemalar oluşturuldu")
}
func GetCollection(databaseName string, collectionName string) *mongo.Collection {
	return MongoClient.Database(databaseName).Collection(collectionName)
}

func CreateUniqueIndexes(databaseName string, collectionName string) error {
	collection := GetCollection(databaseName, collectionName)

	// Email için unique index
	emailIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	// Username için unique index
	usernameIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{emailIndex, usernameIndex})
	return err
}

func createUserCollectionWithSchema() {
	db := MongoClient.Database("authDB")

	userSchema := bson.M{
		"bsonType": "object",
		"required": []string{"username", "email", "password", "createdAt"},
		"properties": bson.M{
			"username": bson.M{
				"bsonType":    "string",
				"minLength":   3,
				"maxLength":   30,
				"description": "must be a string between 3-30 characters",
			},
			"email": bson.M{
				"bsonType":    "string",
				"pattern":     `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
				"description": "must be a valid email address",
			},
			"password": bson.M{
				"bsonType":    "string",
				"minLength":   8,
				"description": "must be at least 8 characters",
			},
			"age": bson.M{
				"bsonType":    "int",
				"minimum":     13,
				"maximum":     150,
				"description": "must be an integer between 13-150",
			},
		},
	}

	cmd := bson.D{
		{Key: "create", Value: "users"},
		{Key: "validator", Value: bson.M{"$jsonSchema": userSchema}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.RunCommand(ctx, cmd).Err(); err != nil {
		fmt.Println("User collection already exists or error:", err)
	}
}

// PasswordReset koleksiyonu ve şeması
func createPasswordResetCollectionWithSchema() {
	db := MongoClient.Database("authDB")

	passwordResetSchema := bson.M{
		"bsonType": "object",
		"required": []string{"userId", "token", "expiresAt", "used"},
		"properties": bson.M{
			"userId": bson.M{
				"bsonType":    "objectId",
				"description": "must be a valid ObjectId reference to users",
			},
			"token": bson.M{
				"bsonType":    "string",
				"minLength":   32,
				"maxLength":   256,
				"description": "must be a valid reset token",
			},
			"expiresAt": bson.M{
				"bsonType":    "date",
				"description": "must be a valid expiration date",
			},
			"used": bson.M{
				"bsonType":    "bool",
				"description": "must be a boolean",
			},
		},
	}

	cmd := bson.D{
		{Key: "create", Value: "passwordresets"},
		{Key: "validator", Value: bson.M{"$jsonSchema": passwordResetSchema}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.RunCommand(ctx, cmd).Err(); err != nil {
		fmt.Println("PasswordReset collection already exists or error:", err)
	}

	// Index oluşturma (Token'a göre hızlı arama için)
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "token", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := db.Collection("passwordresets").Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Printf("PasswordReset index oluşturulamadı: %v", err)
	}
}
