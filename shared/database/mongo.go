package database

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MongoClient      *mongo.Client
	onceMongoConnect sync.Once
)

// MongoDB'ye bağlan
func ConnectMongoDB(mongoURI string) {
	onceMongoConnect.Do(func() {
		fmt.Println("MongoDB bağlantısı başlatılıyor...")

		clientOptions := options.Client().ApplyURI(mongoURI)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			log.Fatalf("MongoDB'ye bağlanılamadı: %v", err)
		}

		// Bağlantıyı test et
		err = client.Ping(ctx, nil)
		if err != nil {
			log.Fatalf("MongoDB ping başarısız: %v", err)
		}
		MongoClient = client
	})
}

// Bir koleksiyona erişim sağlar
func GetCollection(databaseName string, collectionName string) *mongo.Collection {
	if MongoClient == nil {
		log.Fatal("MongoDB bağlantısı henüz başlatılmadı!")
	}
	return MongoClient.Database(databaseName).Collection(collectionName)
}
func GetDatabase(databaseName string) *mongo.Database {
	if MongoClient == nil {
		log.Fatal("MongoDB bağlantısı oluşturulmamış. Önce ConnectMongoDB çağırılmalı!")
	}
	return MongoClient.Database(databaseName)
}

// Bağlantıyı kapat
func DisconnectMongoDB() {
	if MongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := MongoClient.Disconnect(ctx); err != nil {
			log.Fatalf("MongoDB bağlantısı kapatılamadı: %v", err)
		}
		fmt.Println("MongoDB bağlantısı kapatıldı.")
	}
}
