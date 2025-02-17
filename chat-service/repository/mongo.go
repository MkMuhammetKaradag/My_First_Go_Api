package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/MKMuhammetKaradag/go-microservice/shared/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const chatDB = "chatDB"

func InitChatDatabase() {
	CreateUserCollectionWithSchema()
	CreateChatCollectionWithSchema()
	CreateMessageCollectionWithSchema()
	CreateUniqueIndexes()
	fmt.Println("Auth servisinin koleksiyonları oluşturuldu.")
}

func CreateUniqueIndexes() {
	db := database.GetDatabase(chatDB)
	userCollection := db.Collection("users")

	// Mevcut indeksleri al
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	existingIndexes, err := userCollection.Indexes().List(ctx)
	if err != nil {
		fmt.Printf("İndeksleri kontrol ederken hata: %v\n", err)
		return
	}

	// Mevcut indeks adlarını tut
	existingIndexNames := make(map[string]bool)
	for existingIndexes.Next(ctx) {
		var idx bson.M
		if err := existingIndexes.Decode(&idx); err != nil {
			continue
		}
		if name, ok := idx["name"].(string); ok {
			existingIndexNames[name] = true
		}
	}

	// Username indeksini kontrol et ve gerekirse oluştur
	if !existingIndexNames["username_1"] {
		usernameIndexModel := mongo.IndexModel{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		}

		_, err = userCollection.Indexes().CreateOne(ctx, usernameIndexModel)
		if err != nil {
			fmt.Printf("Username index oluşturulurken hata: %v\n", err)
		} else {
			fmt.Println("Username indeksi başarıyla oluşturuldu")
		}
	} else {
		fmt.Println("Username indeksi zaten mevcut")
	}

	// Email indeksini kontrol et ve gerekirse oluştur
	if !existingIndexNames["email_1"] {
		emailIndexModel := mongo.IndexModel{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		}

		_, err = userCollection.Indexes().CreateOne(ctx, emailIndexModel)
		if err != nil {
			fmt.Printf("Email index oluşturulurken hata: %v\n", err)
		} else {
			fmt.Println("Email indeksi başarıyla oluşturuldu")
		}
	} else {
		fmt.Println("Email indeksi zaten mevcut")
	}
}

func CreateUserCollectionWithSchema() {
	db := database.GetDatabase(chatDB)

	// Önce koleksiyonun var olup olmadığını kontrol et
	colNames, err := db.ListCollectionNames(
		context.Background(),
		bson.M{"name": "users"},
	)

	// Koleksiyon yoksa oluştur
	if err == nil && len(colNames) == 0 {
		// Koleksiyon oluşturma seçenekleri
		opts := options.CreateCollection().SetValidator(bson.M{
			"$jsonSchema": bson.M{
				"bsonType": "object",
				"required": []string{"username", "email", "createdAt"},
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
					"age": bson.M{
						"bsonType":    "int",
						"minimum":     13,
						"maximum":     150,
						"description": "must be an integer between 13-150",
					},
				},
			},
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := db.CreateCollection(ctx, "users", opts)
		if err != nil {
			fmt.Println("Error creating users collection:", err)
		} else {
			fmt.Println("Users collection created successfully with schema")
		}
	} else if err != nil {
		fmt.Println("Error checking if collection exists:", err)
	} else {
		fmt.Println("Users collection already exists, skipping creation")
	}
}
func CreateChatCollectionWithSchema() {
	db := database.GetDatabase(chatDB)

	// Önce koleksiyonun var olup olmadığını kontrol et
	colNames, err := db.ListCollectionNames(
		context.Background(),
		bson.M{"name": "chats"},
	)

	// Koleksiyon yoksa oluştur
	if err == nil && len(colNames) == 0 {
		// Koleksiyon oluşturma seçenekleri
		opts := options.CreateCollection().SetValidator(bson.M{
			"$jsonSchema": bson.M{
				"bsonType": "object",
				"required": []string{"chatName", "participants", "createdAt"},
				"properties": bson.M{
					"chatName": bson.M{
						"bsonType":    "string",
						"minLength":   1,
						"maxLength":   100,
						"description": "must be a string between 1-100 characters",
					},
					"participants": bson.M{
						"bsonType":    "array",
						"minItems":    1,
						"uniqueItems": true,
						"items": bson.M{
							"bsonType":    "objectId",
							"description": "must be a valid ObjectId referencing users collection",
						},
						"description": "must be an array of unique user references",
					},
					"admins": bson.M{
						"bsonType":    "array",
						"uniqueItems": true,
						"items": bson.M{
							"bsonType":    "objectId",
							"description": "must be a valid ObjectId referencing users collection",
						},
						"description": "must be an array of unique user references who are admins",
					},
					"createdAt": bson.M{
						"bsonType":    "date",
						"description": "must be a valid date",
					},
					"updatedAt": bson.M{
						"bsonType":    "date",
						"description": "must be a valid date",
					},
				},
			},
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := db.CreateCollection(ctx, "chats", opts)
		if err != nil {
			fmt.Println("Error creating chats collection:", err)
		} else {
			fmt.Println("Chats collection created successfully with schema")
		}
	} else if err != nil {
		fmt.Println("Error checking if chats collection exists:", err)
	} else {
		fmt.Println("Chats collection already exists, skipping creation")
	}
}

func CreateMessageCollectionWithSchema() {
	db := database.GetDatabase(chatDB)

	// Önce koleksiyonun var olup olmadığını kontrol et
	colNames, err := db.ListCollectionNames(
		context.Background(),
		bson.M{"name": "messages"},
	)

	// Koleksiyon yoksa oluştur
	if err == nil && len(colNames) == 0 {
		// Koleksiyon oluşturma seçenekleri
		opts := options.CreateCollection().SetValidator(bson.M{
			"$jsonSchema": bson.M{
				"bsonType": "object",
				"required": []string{"sender", "chat", "content", "createdAt"},
				"properties": bson.M{
					"sender": bson.M{
						"bsonType":    "objectId",
						"description": "must be a valid ObjectId referencing user who sent the message",
					},
					"chat": bson.M{
						"bsonType":    "objectId",
						"description": "must be a valid ObjectId referencing the chat this message belongs to",
					},
					"content": bson.M{
						"bsonType":    "string",
						"description": "must be a string containing the message content",
					},
					"createdAt": bson.M{
						"bsonType":    "date",
						"description": "must be a valid date when message was created",
					},
					"updatedAt": bson.M{
						"bsonType":    "date",
						"description": "must be a valid date when message was last updated",
					},
					"isDeleted": bson.M{
						"bsonType":    "bool",
						"description": "indicates if the message has been deleted",
					},
					"deletedAt": bson.M{
						"bsonType":    "date",
						"description": "must be a valid date when message was deleted, if applicable",
					},
				},
			},
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := db.CreateCollection(ctx, "messages", opts)
		if err != nil {
			fmt.Println("Error creating messages collection:", err)
		} else {
			fmt.Println("Messages collection created successfully with schema")

			// Mesajlar için indeksler oluştur
			messageCollection := db.Collection("messages")

			// Chat ID'ye göre indeks
			chatIndexModel := mongo.IndexModel{
				Keys: bson.D{{Key: "chat", Value: 1}},
			}

			_, err = messageCollection.Indexes().CreateOne(ctx, chatIndexModel)
			if err != nil {
				fmt.Printf("Chat index oluşturulurken hata: %v\n", err)
			}

			// Gönderici ID'ye göre indeks
			senderIndexModel := mongo.IndexModel{
				Keys: bson.D{{Key: "sender", Value: 1}},
			}

			_, err = messageCollection.Indexes().CreateOne(ctx, senderIndexModel)
			if err != nil {
				fmt.Printf("Sender index oluşturulurken hata: %v\n", err)
			}

			// Oluşturulma tarihine göre indeks (zaman sıralaması için)
			createdAtIndexModel := mongo.IndexModel{
				Keys: bson.D{{Key: "createdAt", Value: 1}},
			}

			_, err = messageCollection.Indexes().CreateOne(ctx, createdAtIndexModel)
			if err != nil {
				fmt.Printf("CreatedAt index oluşturulurken hata: %v\n", err)
			}
		}
	} else if err != nil {
		fmt.Println("Error checking if messages collection exists:", err)
	} else {
		fmt.Println("Messages collection already exists, skipping creation")
	}
}
