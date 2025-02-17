package routes

import (
	"encoding/json"
	"net/http"

	"github.com/MKMuhammetKaradag/go-microservice/chat-service/controllers"
	"github.com/MKMuhammetKaradag/go-microservice/shared/messaging"
	"github.com/MKMuhammetKaradag/go-microservice/shared/middlewares"
	"github.com/MKMuhammetKaradag/go-microservice/shared/redisrepo"
	"github.com/go-chi/chi/v5"
)

func CreateServer(rabbitMQ *messaging.RabbitMQ, sessionRepo *redisrepo.RedisRepository) *chi.Mux {
	chatController := controllers.NewChatController(rabbitMQ, sessionRepo)
	authMiddleware := middlewares.NewAuthMiddleware(sessionRepo)
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
			protectedRouter.Post("/message/create", chatController.SendMessage)
			protectedRouter.Post("/getUsers", chatController.GetChatUsers)

		})
	})
	// r.Get("/auth/ws", wsController.HandleWebSocket)
	return r
}
