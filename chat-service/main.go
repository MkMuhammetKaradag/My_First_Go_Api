package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MKMuhammetKaradag/go-microservice/chat-service/repository"
	"github.com/MKMuhammetKaradag/go-microservice/chat-service/routes"
	"github.com/MKMuhammetKaradag/go-microservice/shared/database"
	"github.com/MKMuhammetKaradag/go-microservice/shared/messaging"
	"github.com/MKMuhammetKaradag/go-microservice/shared/models"
	"github.com/MKMuhammetKaradag/go-microservice/shared/redisrepo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// @title           Chat Service API
// @version         1.0
// @description     This is an Chat service for managing chat.
// @host            localhost:8083
// @BasePath       /
func main() {

	database.ConnectMongoDB("mongodb://localhost:27017/chatDB")

	repository.InitChatDatabase()
	database.ConnectRedis("localhost:6379", 0)
	chatRepo := repository.NewChatRepository(database.GetCollection("chatDB", "chats"))
	// a, error1 := chatRepo.IsUserInChat("", "")
	// fmt.Println(error1)
	// fmt.Println(a)

	config := messaging.NewDefaultConfig()
	config.RetryTypes = []string{"user_created"}
	redisRepo := redisrepo.NewRedisRepository(database.RedisClient) // Redis repository oluşturuldu
	var err error
	rabbitMQ, err := messaging.NewRabbitMQ(config, messaging.ChatService)
	if err != nil {
		log.Fatal("RabbitMQ bağlantı hatası:", err)
	}
	defer rabbitMQ.Close()
	err = rabbitMQ.ConsumeMessages(func(msg messaging.Message) error {
		fmt.Println(msg.Type)
		if msg.Type == "user_created" {
			fmt.Println("user_creat geldi")
			fmt.Println(msg)
			return handleUserCreated(msg)
		}
		return nil
	})
	port := 8083

	// Servisi başlat
	// http.ListenAndServe(":8083", nil)
	fmt.Printf("chat Service running on port %d\n", port)
	r := routes.CreateServer(rabbitMQ, chatRepo, redisRepo)
	http.ListenAndServe(fmt.Sprintf(":%d", port), r)

}
func handleUserCreated(msg messaging.Message) error {
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("geçersiz mesaj formatı")
	}

	// User servisindeki MongoDB'ye kaydet
	objectID, err := primitive.ObjectIDFromHex(data["user_id"].(string))
	if err != nil {
		fmt.Println("Error:", err)
		return fmt.Errorf("kullanıcı oluşturma hatası: %v", err)
	}

	ageValue, ok := data["age"].(float64)
	if !ok {
		// Hata durumunu işle
		panic("Invalid type for age")
	}

	ageInt := int(ageValue) // float64'ü int'e dönüştür
	agePointer := &ageInt   // pointer'a dönüştür

	user := models.User{
		ID:        objectID,
		Email:     data["email"].(string),
		FirstName: data["firstName"].(string),
		Username:  data["username"].(string),
		Age:       agePointer,
	}

	err = repository.CreateUser(&user)
	if err != nil {
		return fmt.Errorf("kullanıcı oluşturma hatası: %v", err)
	}

	log.Printf("Yeni kullanıcı oluşturuldu: %s", user.Email)
	return nil
}
