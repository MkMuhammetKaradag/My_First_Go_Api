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
}
func GetCollection(databaseName string,  collectionName string ) *mongo.Collection {
	return MongoClient.Database(databaseName).Collection(collectionName)
}

func CreateUniqueIndexes( databaseName string,  collectionName string) error {
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