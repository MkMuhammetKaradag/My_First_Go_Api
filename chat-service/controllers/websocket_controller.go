// controllers/websocket_controller.go
package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MKMuhammetKaradag/go-microservice/chat-service/repository"
	myWebsocket "github.com/MKMuhammetKaradag/go-microservice/chat-service/websocket"
	"github.com/MKMuhammetKaradag/go-microservice/shared/middlewares"
	"github.com/MKMuhammetKaradag/go-microservice/shared/redisrepo"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WebSocketController struct {
	Hub *myWebsocket.Hub

	RedisRepo      *redisrepo.RedisRepository
	chatRepository *repository.ChatRepository
}

func NewWebSocketController(hub *myWebsocket.Hub, chatRepo *repository.ChatRepository, redisRepo *redisrepo.RedisRepository) *WebSocketController {
	return &WebSocketController{
		Hub:            hub,
		chatRepository: chatRepo,
		RedisRepo:      redisRepo,
	}
}

func (wc *WebSocketController) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userData, ok := middlewares.GetUserData(r)
	// fmt.Println("hello", userData)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Kullanıcı bilgisi bulunamadı")
		return
	}
	userID, exists := userData["id"]
	if !exists {
		fmt.Println("id not found in userData")
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	chatID := chi.URLParam(r, "chatID")

	if chatID == "" {
		log.Println("Chat ID is required")
		conn.Close()
		return
	}

	isMember, err := wc.chatRepository.IsUserInChat(chatID, userID)
	if err != nil {
		fmt.Println("Hata:", err.Error())
		message := websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Server Error")
		conn.WriteMessage(websocket.CloseMessage, message) // Kapatma mesajı gönder
		conn.Close()                                       // WebSocket bağlantısını kapat
		return
	} else if !isMember {
		fmt.Println("Kullanıcı bu sohbette yok.")
		respondWithError(w, http.StatusForbidden, "User does not have permission to listen to the chat")
		message := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Unauthorized access")
		conn.WriteMessage(websocket.CloseMessage, message) // Kapatma mesajı gönder
		conn.Close()                                       // WebSocket bağlantısını kapat
		return
	}

	client := &myWebsocket.Client{
		ChatID: chatID,
		Conn:   conn,
	}

	wc.Hub.Register <- client

	defer func() {
		wc.Hub.Unregister <- client
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}

}
