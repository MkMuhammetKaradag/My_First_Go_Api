package routes

import (
	"encoding/json"
	"net/http"

	"github.com/MKMuhammetKaradag/go-microservice/chat-service/controllers"
	"github.com/MKMuhammetKaradag/go-microservice/chat-service/repository"
	"github.com/MKMuhammetKaradag/go-microservice/chat-service/websocket"
	"github.com/MKMuhammetKaradag/go-microservice/shared/messaging"
	"github.com/MKMuhammetKaradag/go-microservice/shared/middlewares"
	"github.com/MKMuhammetKaradag/go-microservice/shared/redisrepo"
	"github.com/go-chi/chi/v5"
)

func CreateServer(rabbitMQ *messaging.RabbitMQ, chatRepo *repository.ChatRepository, sessionRepo *redisrepo.RedisRepository) *chi.Mux {
	chatController := controllers.NewChatController(rabbitMQ, sessionRepo)
	authMiddleware := middlewares.NewAuthMiddleware(sessionRepo)
	hub := websocket.NewHub()
	go hub.Run()
	go hub.ListenRedisSendMessage(sessionRepo)
	wsController := controllers.NewWebSocketController(hub, chatRepo, sessionRepo)
	r := chi.NewRouter()
	r.Use(middlewares.Logger)
	r.Route("/chat", func(r chi.Router) {
		r.Get("/chat", func(w http.ResponseWriter, r *http.Request) {

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"message": "get Chat",
				"chat":    "chat",
			})

		})

		r.Group(func(protectedRouter chi.Router) {
			protectedRouter.Use(authMiddleware.Authenticate)
			protectedRouter.Post("/create", chatController.CreateChat)
			protectedRouter.Get("/{chatID}", chatController.CreateChat)
			protectedRouter.Get("/myChats", chatController.GetMyChats)
			protectedRouter.Post("/message/create", chatController.SendMessage)
			protectedRouter.Post("/addParticipants", chatController.AddParticipants)
			protectedRouter.Post("/leave/{chatID}", chatController.LeaveChat)
			protectedRouter.Get("/chatDetail", chatController.GetChatUsers)
			protectedRouter.Get("/chatlisten/{chatID}", wsController.HandleWebSocket)
		})
	})

	return r
}
