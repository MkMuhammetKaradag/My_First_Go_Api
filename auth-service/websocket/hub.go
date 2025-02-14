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
	UserID string
	Conn   *websocket.Conn
}

type Hub struct {
	Clients    map[string]*Client
	Register   chan *Client
	Unregister chan *Client
	Mutex      sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[string]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Mutex.Lock()
			h.Clients[client.UserID] = client
			h.Mutex.Unlock()
		case client := <-h.Unregister:
			h.Mutex.Lock()
			delete(h.Clients, client.UserID)
			h.Mutex.Unlock()
			client.Conn.Close()
		}
	}
}

func (h *Hub) ListenRedisStatus(redisRepo *redisrepo.RedisRepository) {
	pubsub := redisRepo.Client.Subscribe("user_status")
	defer pubsub.Close()

	for {
		msg, err := pubsub.ReceiveMessage()
		if err != nil {
			log.Println("Redis sub error:", err)
			continue
		}

		parts := strings.Split(msg.Payload, ":")
		if len(parts) != 2 {
			continue
		}
		userID, status := parts[0], parts[1]

		h.Mutex.RLock()
		if client, ok := h.Clients[userID]; ok {
			client.Conn.WriteJSON(map[string]string{
				"event":  "status_update",
				"userID": userID,
				"status": status,
			})
		}
		h.Mutex.RUnlock()
	}
}
