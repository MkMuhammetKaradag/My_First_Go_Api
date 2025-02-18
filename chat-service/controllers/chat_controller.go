package controllers

import (
	"encoding/json"

	"net/http"

	"github.com/MKMuhammetKaradag/go-microservice/chat-service/dto"
	"github.com/MKMuhammetKaradag/go-microservice/chat-service/services"
	"github.com/MKMuhammetKaradag/go-microservice/shared/messaging"
	"github.com/MKMuhammetKaradag/go-microservice/shared/middlewares"
	"github.com/MKMuhammetKaradag/go-microservice/shared/models"
	"github.com/MKMuhammetKaradag/go-microservice/shared/redisrepo"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatController struct {
	chatService *services.ChatService
	rabbitMQ    *messaging.RabbitMQ
	sessionRepo *redisrepo.RedisRepository
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
func NewChatController(rabbitMQ *messaging.RabbitMQ, sessionRepo *redisrepo.RedisRepository) *ChatController {
	return &ChatController{
		chatService: services.NewChatService(),
		rabbitMQ:    rabbitMQ,
		sessionRepo: sessionRepo,
	}
}

func (ctrl *ChatController) CreateChat(w http.ResponseWriter, r *http.Request) {
	var input models.Chat
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Geçersiz veri")
		return
	}

	userData, ok := middlewares.GetUserData(r)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Kullanıcı bilgisi bulunamadı")
		return
	}

	userID, err := primitive.ObjectIDFromHex(userData["id"])
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Geçersiz kullanıcı ID formatı")
		return
	}

	// Admins dizisini oluştur veya mevcut diziye ekle
	if input.Admins == nil {
		input.Admins = []primitive.ObjectID{userID}
	} else {
		input.Admins = append(input.Admins, userID)
	}

	// participantExists := false
	// for _, participant := range input.Participants {
	// 	if participant == userID {
	// 		participantExists = true
	// 		break
	// 	}
	// }
	// if !participantExists {
	// 	input.Participants = append(input.Participants, userID)
	// }
	// input.Admins = []
	chat, err := ctrl.chatService.CreateChat(&input)
	if err != nil {
		respondWithError(w, http.StatusConflict, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{
		"message": "chat  başarıyla oluşturuldu",
		"chat":    chat,
	})

}

func (ctrl *ChatController) SendMessage(w http.ResponseWriter, r *http.Request) {

	var input models.Message
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Geçersiz veri")
		return
	}

	userData, ok := middlewares.GetUserData(r)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Kullanıcı bilgisi bulunamadı")
		return
	}

	userID, err := primitive.ObjectIDFromHex(userData["id"])
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Geçersiz kullanıcı ID formatı")
		return
	}

	input.Sender = userID
	// Admins dizisini oluştur veya mevcut diziye ekle
	message, err := ctrl.chatService.SendMessage(&input)
	if err != nil {
		respondWithError(w, http.StatusConflict, err.Error())
		return
	}
	ctrl.sessionRepo.PublishChatMessage(string(input.Chat.Hex()), message.Content, string(userID.Hex()))
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{
		"message":     "message  başarıyla oluşturuldu",
		"chatMessage": message,
	})
}

func (ctrl *ChatController) GetChatUsers(w http.ResponseWriter, r *http.Request) {

	var input dto.GetChatUsersDto
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Geçersiz veri")
		return
	}

	// Admins dizisini oluştur veya mevcut diziye ekle
	chat, err := ctrl.chatService.GetChatWithUsersAggregation(input.ChatID)
	if err != nil {
		respondWithError(w, http.StatusConflict, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{
		"message": "chat  başarıyla çekildi",
		"chat":    chat,
	})
}

func (ctrl *ChatController) GetMyChats(w http.ResponseWriter, r *http.Request) {
	userData, ok := middlewares.GetUserData(r)
	// fmt.Println("hello", userData)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Kullanıcı bilgisi bulunamadı")
		return
	}
	userID, exists := userData["id"]
	if !exists {
		respondWithError(w, http.StatusInternalServerError, "Kullanıcı bilgisi bulunamadı")
		return
	}
	// Admins dizisini oluştur veya mevcut diziye ekle
	chat, err := ctrl.chatService.GetMyChatsWithUsersAggregation(userID)
	if err != nil {
		respondWithError(w, http.StatusConflict, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{
		"message": "chat  başarıyla çekildi",
		"chat":    chat,
	})

}

func (ctrl *ChatController) AddParticipants(w http.ResponseWriter, r *http.Request) {
	userData, ok := middlewares.GetUserData(r)
	// fmt.Println("hello", userData)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Kullanıcı bilgisi bulunamadı")
		return
	}
	userID, exists := userData["id"]
	if !exists {
		respondWithError(w, http.StatusInternalServerError, "Kullanıcı bilgisi bulunamadı")
		return
	}

	var input dto.ChatAddParticipants
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Geçersiz veri")
		return
	}
	chat, err := ctrl.chatService.AddParticipants(userID, &input)
	if err != nil {
		respondWithError(w, http.StatusConflict, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{
		"message": "chat  başarıyla  katılımcı eklendi ",
		"chat":    chat,
	})
}
func (ctrl *ChatController) LeaveChat(w http.ResponseWriter, r *http.Request) {
	userData, ok := middlewares.GetUserData(r)
	// fmt.Println("hello", userData)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Kullanıcı bilgisi bulunamadı")
		return
	}
	userID, exists := userData["id"]
	if !exists {
		respondWithError(w, http.StatusInternalServerError, "Kullanıcı bilgisi bulunamadı")
		return
	}
	chatID := chi.URLParam(r, "chatID")

	if chatID == "" {
		respondWithError(w, http.StatusInternalServerError, "Chat bilgisi bulunamadı")
		return
	}
	chat, err := ctrl.chatService.LeaveChat(userID,chatID)
	if err != nil {
		respondWithError(w, http.StatusConflict, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{
		"message": "chat  başarıyla  katılımcı eklendi ",
		"chat":    chat,
	})
}
