package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MKMuhammetKaradag/go-microservice/auth-service/database"
	"github.com/MKMuhammetKaradag/go-microservice/auth-service/routes"
)

func main() {
	// Veritabanlarına bağlan
	database.ConnectMongoDB("mongodb://localhost:27017/authDB")
	if err := database.CreateUniqueIndexes("authDB", "users"); err != nil {
		log.Fatal("Index oluşturulurken hata:", err)
	}
	database.ConnectRedis()

	port := 8080
	fmt.Printf("Auth Service running on port %d\n", port)
	r := routes.CreateServer()
	http.ListenAndServe(fmt.Sprintf(":%d", port), r)

}
