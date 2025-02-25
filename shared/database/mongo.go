package database

import (
	"context"
	"fmt"
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
func ConnectMongoDB(mongoURI string) error {
	var err error

	onceMongoConnect.Do(func() {
		fmt.Println("MongoDB bağlantısı başlatılıyor...")

		clientOptions := options.Client().ApplyURI(mongoURI)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, connErr := mongo.Connect(ctx, clientOptions)
		if connErr != nil {
			err = fmt.Errorf("MongoDB bağlantı hatası: %w", connErr)
			return
		}

		// Bağlantıyı test et
		if pingErr := client.Ping(ctx, nil); pingErr != nil {
			err = fmt.Errorf("MongoDB ping başarısız: %w", pingErr)
			return
		}

		MongoClient = client
	})

	return err
}

// Bir koleksiyona erişim sağlar
func GetCollection(databaseName string, collectionName string) (*mongo.Collection, error) {
	if MongoClient == nil {
		return nil, fmt.Errorf("MongoDB bağlantısı henüz başlatılmadı")
	}
	return MongoClient.Database(databaseName).Collection(collectionName), nil
}

// Veritabanına erişim sağlar
func GetDatabase(databaseName string) (*mongo.Database, error) {
	if MongoClient == nil {
		return nil, fmt.Errorf("MongoDB bağlantısı oluşturulmamış, önce ConnectMongoDB çağırılmalı")
	}
	return MongoClient.Database(databaseName), nil
}

// Bağlantıyı kapat
func DisconnectMongoDB() error {
	if MongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := MongoClient.Disconnect(ctx); err != nil {
			return fmt.Errorf("MongoDB bağlantısı kapatılamadı: %w", err)
		}
		fmt.Println("MongoDB bağlantısı kapatıldı.")
	}
	return nil
}
