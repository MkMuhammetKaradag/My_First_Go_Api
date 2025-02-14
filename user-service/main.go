package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MKMuhammetKaradag/go-microservice/shared/database"
	"github.com/MKMuhammetKaradag/go-microservice/shared/messaging"
	"github.com/MKMuhammetKaradag/go-microservice/shared/models"
	"github.com/MKMuhammetKaradag/go-microservice/shared/redisrepo"
	"github.com/MKMuhammetKaradag/go-microservice/user-service/repository"
	"github.com/MKMuhammetKaradag/go-microservice/user-service/routes"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	// Veritabanlarına bağlan
	database.ConnectMongoDB("mongodb://localhost:27017/userDB")
	if err := repository.CreateUniqueIndexes("userDB", "users"); err != nil {
		log.Fatal("Index oluşturulurken hata:", err)
	}
	repository.InitUserDatabase()
	// database.ConnectRedis()
	database.ConnectRedis("localhost:6379", 0)
	config := messaging.NewDefaultConfig()
	config.RetryTypes = []string{"user_created"}
	redisRepo := redisrepo.NewRedisRepository(database.RedisClient) // Redis repository oluşturuldu
	rabbit, err := messaging.NewRabbitMQ(config, messaging.UserService)
	if err != nil {
		log.Fatal("RabbitMQ bağlantı hatası:", err)
	}
	defer rabbit.Close()

	// Mesaj dinleyiciyi başlat
	err = rabbit.ConsumeMessages(func(msg messaging.Message) error {
		fmt.Println(msg.Type)
		if msg.Type == "user_created" {
			fmt.Println("user_creat geldi")
			fmt.Println(msg)
			return handleUserCreated(msg)
		}
		return nil
	})
	if err != nil {
		log.Fatal("Mesaj dinleyici başlatılamadı:", err)
	}
	port := 8081
	fmt.Printf("Auth Service running on port %d\n", port)
	r := routes.CreateServer(redisRepo)
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
