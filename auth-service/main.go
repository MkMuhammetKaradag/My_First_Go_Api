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

// @title           Authentication Service API
// @version         1.0
// @description     This is an authentication service for managing user authentication and authorization.
// @host           localhost:8080
// @BasePath       /
func main() {
	// Servisi başlat
	if err := startAuthService(); err != nil {
		log.Fatal("Servis başlatılamadı:", err)
	}
}

// Auth servisini başlatan fonksiyon
func startAuthService() error {
	// Veritabanı bağlantılarını başlat
	if err := initDatabases(); err != nil {
		return fmt.Errorf("veritabanı hatası: %w", err)
	}

	// RabbitMQ bağlantısını oluştur
	rabbitMQ, err := initRabbitMQ()
	if err != nil {
		return fmt.Errorf("RabbitMQ hatası: %w", err)
	}
	defer rabbitMQ.Close()

	// Sunucuyu başlat
	return startServer(rabbitMQ)
}

// Veritabanı bağlantılarını başlatan fonksiyon
func initDatabases() error {
	// MongoDB bağlantısı
	if err := database.ConnectMongoDB("mongodb://localhost:27017/authDB"); err != nil {
		return fmt.Errorf("MongoDB bağlantı hatası: %w", err)
	}

	// Kullanıcı koleksiyonunda indeksleri oluştur
	if err := repository.CreateUniqueIndexes("authDB", "users"); err != nil {
		return fmt.Errorf("MongoDB indeks hatası: %w", err)
	}

	// Auth veritabanını başlat
	repository.InitAuthDatabase()

	// Redis bağlantısını kur
	if err := database.ConnectRedis("localhost:6379", 0); err != nil {
		return fmt.Errorf("Redis bağlantı hatası: %w", err)
	}
	if !database.IsRedisConnected() {
		return fmt.Errorf("Redis bağlantısı aktif değil!")
	}

	// Uygulama kapanırken Redis bağlantısını kapat
	// defer func() {
	// 	if err := database.DisconnectRedis(); err != nil {
	// 		log.Printf("Redis bağlantısı kapatılırken hata: %v", err)
	// 	}
	// }()
	return nil
}

// RabbitMQ bağlantısını başlatan fonksiyon
func initRabbitMQ() (*messaging.RabbitMQ, error) {
	config := messaging.NewDefaultConfig()
	config.RetryTypes = []string{"user_created"}

	rabbitMQ, err := messaging.NewRabbitMQ(config, messaging.AuthService)
	if err != nil {
		log.Printf("RabbitMQ bağlantısı kurulamadı: %v", err)
		return nil, err
	}

	return rabbitMQ, nil
}

// HTTP sunucusunu başlatan fonksiyon
func startServer(rabbitMQ *messaging.RabbitMQ) error {
	port := 8080
	fmt.Printf("Auth Service running on port %d\n", port)

	collection, _ := database.GetCollection("authDB", "users")
	userRepo := repository.NewUserRepository(collection)
	redisRepo := redisrepo.NewRedisRepository(database.RedisClient)

	// Router oluştur
	r := routes.CreateServer(rabbitMQ, redisRepo, userRepo)

	// HTTP sunucusunu başlat
	return http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
