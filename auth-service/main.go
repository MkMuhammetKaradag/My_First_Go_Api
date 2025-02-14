package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MKMuhammetKaradag/go-microservice/auth-service/repository"
	"github.com/MKMuhammetKaradag/go-microservice/auth-service/routes"
	"github.com/MKMuhammetKaradag/go-microservice/shared/database"
	"github.com/MKMuhammetKaradag/go-microservice/shared/messaging"
	"github.com/MKMuhammetKaradag/go-microservice/shared/redisrepo"
)

func main() {
	// Veritabanlarına bağlan
	database.ConnectMongoDB("mongodb://localhost:27017/authDB")
	if err := repository.CreateUniqueIndexes("authDB", "users"); err != nil {
		log.Fatal("Index oluşturulurken hata:", err)
	}
	repository.InitAuthDatabase()
	// database.ConnectRedis()
	database.ConnectRedis("localhost:6379", 0)
	userRepo := repository.NewUserRepository(database.GetCollection("authDB", "users"))

	config := messaging.NewDefaultConfig()
	config.RetryTypes = []string{"user_created"}
	redisRepo := redisrepo.NewRedisRepository(database.RedisClient) // Redis repository oluşturuldu
	var err error
	rabbitMQ, err := messaging.NewRabbitMQ(config, messaging.AuthService)
	if err != nil {
		log.Fatal("RabbitMQ bağlantı hatası:", err)
	}
	defer rabbitMQ.Close()

	port := 8080
	fmt.Printf("Auth Service running on port %d\n", port)
	r := routes.CreateServer(rabbitMQ, redisRepo, userRepo)
	http.ListenAndServe(fmt.Sprintf(":%d", port), r)

}
