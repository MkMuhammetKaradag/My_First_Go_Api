// controllers/websocket_controller.go
package controllers

import (
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
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Kullanıcı bilgisi bulunamadı")
		return
	}

	userID, exists := userData["id"]
	if !exists {
		respondWithError(w, http.StatusInternalServerError, "User ID bulunamadı")
		return
	}

	// WebSocket bağlantısını yükselt
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		respondWithError(w, http.StatusInternalServerError, "WebSocket bağlantısı kurulamıyor")
		return
	}

	// Client'i hub'a kaydet
	client := &myWebsocket.Client{UserID: userID, Conn: conn}
	wc.Hub.Register <- client

	// Kullanıcıyı online olarak işaretle
	wc.setUserStatus(userID, "online")
	defer wc.setUserStatus(userID, "offline")

	// Bağlantı kapatıldığında kullanıcıyı offline yap
	defer func() {
		wc.Hub.Unregister <- client
		conn.Close()
	}()

	// Mesaj dinleme döngüsü
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket message error:", err)
			break
		}
	}
}

// Kullanıcı durumu güncelleme ve Redis ile senkronize etme
func (wc *WebSocketController) setUserStatus(userID, status string) {
	// Kullanıcı durumunu güncelle
	err := wc.UserRepo.UpdateUserStatus(userID, status)
	if err != nil {
		log.Printf("Kullanıcı durumu güncellenemedi: %v", err)
	}

	// Redis'e durumu yayınla
	err = wc.RedisRepo.PublishStatus(userID, status)
	if err != nil {
		log.Printf("Redis durumu güncellenemedi: %v", err)
	}
}
