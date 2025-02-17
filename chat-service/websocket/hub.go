// websocket/hub.go
package websocket

import (
	"log"
	"strings"
	"sync"

	"github.com/MKMuhammetKaradag/go-microservice/shared/redisrepo"
	"github.com/gorilla/websocket"
)

type Client struct {
	ChatID string
	Conn   *websocket.Conn
}

type Hub struct {
	Clients    map[string]map[*websocket.Conn]bool // Her chat ID'si için birden fazla istemci
	Register   chan *Client
	Unregister chan *Client
	Mutex      sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[string]map[*websocket.Conn]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Mutex.Lock()
			if _, ok := h.Clients[client.ChatID]; !ok {
				h.Clients[client.ChatID] = make(map[*websocket.Conn]bool)
			}
			h.Clients[client.ChatID][client.Conn] = true
			h.Mutex.Unlock()

		case client := <-h.Unregister:
			h.Mutex.Lock()
			if _, ok := h.Clients[client.ChatID]; ok {
				delete(h.Clients[client.ChatID], client.Conn)
				if len(h.Clients[client.ChatID]) == 0 {
					delete(h.Clients, client.ChatID)
				}
				client.Conn.Close()
			}
			h.Mutex.Unlock()
		}
	}
}

func (h *Hub) ListenRedisSendMessage(redisRepo *redisrepo.RedisRepository) {
	pubsub := redisRepo.Client.Subscribe("send_Message")
	defer pubsub.Close()

	for {
		msg, err := pubsub.ReceiveMessage()
		if err != nil {
			log.Println("Redis sub error:", err)
			continue
		}

		parts := strings.Split(msg.Payload, ":")
		if len(parts) != 3 {
			continue
		}
		chatID, content, senderID := parts[0], parts[1], parts[2]

		h.Mutex.RLock()
		if clients, ok := h.Clients[chatID]; ok {
			for clientConn := range clients {
				err := clientConn.WriteJSON(map[string]string{
					"event":    "send_Message",
					"chatID":   chatID,
					"content":  content,
					"senderID": senderID,
				})
				if err != nil {
					log.Println("WebSocket write error:", err)
					clientConn.Close()
					delete(h.Clients[chatID], clientConn)
				}
			}
		}
		h.Mutex.RUnlock()
	}
}
