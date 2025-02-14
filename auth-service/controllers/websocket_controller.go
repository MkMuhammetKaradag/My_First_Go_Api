// controllers/websocket_controller.go
package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MKMuhammetKaradag/go-microservice/auth-service/repository"
	myWebsocket "github.com/MKMuhammetKaradag/go-microservice/auth-service/websocket"
	"github.com/MKMuhammetKaradag/go-microservice/shared/middlewares"
	"github.com/MKMuhammetKaradag/go-microservice/shared/redisrepo"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WebSocketController struct {
	Hub       *myWebsocket.Hub
	UserRepo  *repository.UserRepository
	RedisRepo *redisrepo.RedisRepository
}

func NewWebSocketController(hub *myWebsocket.Hub, userRepo *repository.UserRepository, redisRepo *redisrepo.RedisRepository) *WebSocketController {
	return &WebSocketController{
		Hub:       hub,
		UserRepo:  userRepo,
		RedisRepo: redisRepo,
	}
}

func (wc *WebSocketController) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	
	userData, ok := middlewares.GetUserData(r)
	fmt.Println("hello", userData)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Kullanıcı bilgisi bulunamadı")
		return
	}

	userID, exists := userData["id"]
	if !exists {
		fmt.Println("id not found in userData")
	} else {
		fmt.Println("User ID:", userID)
	}
	fmt.Println("websovcket içi")
	// userID := "asdsds"
	// WebSocket bağlantısını yükselt
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	// Client'i hub'a kaydet
	client := &myWebsocket.Client{UserID: userID, Conn: conn}
	wc.Hub.Register <- client
	defer func() {
		wc.Hub.Unregister <- client
	}()

	// Kullanıcıyı online olarak işaretle
	wc.UserRepo.UpdateUserStatus(userID, "online")
	wc.RedisRepo.PublishStatus(userID, "online")

	// Bağlantı kapatıldığında offline yap
	defer func() {
		wc.UserRepo.UpdateUserStatus(userID, "offline")
		wc.RedisRepo.PublishStatus(userID, "offline")
	}()

	// Mesaj dinleme döngüsü
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
