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
	database.ConnectMongoDB("mongodb://localhost:27017")
	if err := database.CreateUniqueIndexes("authDB","users"); err != nil {
        log.Fatal("Index oluşturulurken hata:", err)
    }
	database.ConnectRedis()

	// r := gin.Default()

	// // Auth Routes
	// r.POST("/register", routes.Register)
	// r.POST("/login", routes.Login)
	// r.POST("/logout", authMiddleware.AuthMiddleware,routes.Logout)
	// r.GET("/protected", authMiddleware.AuthMiddleware, func(c *gin.Context) {

	// 	userData, exists := c.Get("userData")
	// 	if !exists {
	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kullanıcı verisi mevcut değil"})
	// 		return
	// 	}
	
	// 	// UserData'yı doğru bir şekilde dönüştürerek geri gönderiyoruz
	// 	// Burada userData, map[string]string türünde olduğu için direkt olarak döndürebiliriz.
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"message":   "Bu korumalı bir alan",
	// 		"userData":  userData,  // Kullanıcı bilgilerini burada ekliyoruz
	// 	})

	// 	c.JSON(200, gin.H{"message": "Bu korumalı bir alan"})
	// })

	port := 8080
	fmt.Printf("Auth Service running on port %d\n", port)
	r := routes.CreateServer()
	http.ListenAndServe(fmt.Sprintf(":%d", port), r)
	// log.Fatal(r.Run(fmt.Sprintf(":%d", port)))
}
