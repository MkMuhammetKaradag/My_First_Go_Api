package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/MKMuhammetKaradag/go-microservice/auth-service/dto"
	"github.com/MKMuhammetKaradag/go-microservice/auth-service/services"
	"github.com/MKMuhammetKaradag/go-microservice/shared/database"
	"github.com/MKMuhammetKaradag/go-microservice/shared/messaging"
	"github.com/MKMuhammetKaradag/go-microservice/shared/middlewares"
	"github.com/MKMuhammetKaradag/go-microservice/shared/models"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type AuthController struct {
	authService *services.AuthService
	rabbitMQ    *messaging.RabbitMQ
}

func NewAuthController(rabbitMQ *messaging.RabbitMQ) *AuthController {
	return &AuthController{
		authService: services.NewAuthService(),
		rabbitMQ:    rabbitMQ,
	}
}

var validate = validator.New()

func validateUser(user *models.User) error {
	return validate.Struct(user)
}
func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (ctrl *AuthController) SignUp(w http.ResponseWriter, r *http.Request) {
	var user = models.NewUser()

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		respondWithError(w, http.StatusBadRequest, "Geçersiz veri")
		return
	}

	if err := validateUser(&user); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	activationCode, activationToken, err := ctrl.authService.SignUp(&user)

	if err != nil {
		respondWithError(w, http.StatusConflict, err.Error())
		return
	}
	userActiveMessage := messaging.Message{
		Type:      "active_user",
		ToService: messaging.EmailService,
		Data: map[string]interface{}{
			"email":           user.Email,
			"activation_code": activationCode,
			"template_name":   "activation_email.html",
			"userName":        user.Username,
		},
	}

	err = ctrl.rabbitMQ.PublishMessage(context.Background(), userActiveMessage)
	if err != nil {
		log.Printf("Kullanıcı oluşturma mesajı gönderilemedi: %v", err)

	}
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{
		"message":             "Kullanıcı başarıyla oluşturuldu-asa",
		"userActivationToken": activationToken,
	})

}

func (ctrl *AuthController) ActivationUser(w http.ResponseWriter, r *http.Request) {

	var activationRequest dto.ActivationRequest
	if err := json.NewDecoder(r.Body).Decode(&activationRequest); err != nil {
		respondWithError(w, http.StatusBadRequest, "Geçersiz veri")
		return
	}
	activationUser, err := ctrl.authService.ActivationUser(activationRequest.ActivationCode, activationRequest.ActivationToken)
	if err != nil {
		respondWithError(w, http.StatusConflict, err.Error())
		return
	}
	userCreatedMessage := messaging.Message{
		Type:      "user_created",
		ToService: messaging.ServiceType("UserService"),
		Data: map[string]interface{}{
			"user_id":   activationUser.ID, // MongoDB'de oluşan ID
			"email":     activationUser.Email,
			"firstName": activationUser.FirstName,
			"age":       activationUser.Age,
			"createdAt": activationUser.CreatedAt,
			"username":  activationUser.Username,
			// Diğer gerekli kullanıcı bilgileri
		},
	}

	err = ctrl.rabbitMQ.PublishMessage(context.Background(), userCreatedMessage)
	if err != nil {
		log.Printf("Kullanıcı oluşturma mesajı gönderilemedi: %v", err)
		// İşleme devam et, kritik hata değil
	}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{
		"message": "Kullanıcı başarıyla oluşturuldu-asa",
		"user":    activationUser,
	})
}

func (ctrl *AuthController) SignIn(w http.ResponseWriter, r *http.Request) {
	var input models.User
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Geçersiz veri")
		return
	}

	user, err := ctrl.authService.SignIn(&input)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    user.ID,
		Path:     "/",
		MaxAge:   60 * 60 * 24, // 30 dakika
		HttpOnly: true,
		Secure:   false, // HTTPS kullanıyorsanız true yapın
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"message": "Giriş başarılı",
		"user":    user,
	})
}

func (ctrl *AuthController) Logout(w http.ResponseWriter, r *http.Request) {

	cookieSessionId, err := r.Cookie("session_id")

	if err != nil {

		respondWithError(w, http.StatusInternalServerError, "Giriş yapılmamış")

		return
	}

	sessionKey := "session:" + cookieSessionId.Value
	err = database.RedisClient.Del(sessionKey).Err()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Oturum sonlandırılamadı")
		return
	}
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]string{
		"message": "Başarıyla çıkış yapıldı",
	})
}

func Protected(w http.ResponseWriter, r *http.Request) {

	userData, ok := middlewares.GetUserData(r)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Kullanıcı bilgisi bulunamadı")
		return
	}

	fmt.Println(userData)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Protected endpoint",
		"user":    userData["username"],
	})
}

func (ctrl *AuthController) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var input dto.ForgotPasswordDto
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Geçersiz veri")
		return
	}

	token, userName, err := ctrl.authService.ForgotPassword(input.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	userActiveMessage := messaging.Message{
		Type:      "forgot_password",
		ToService: messaging.EmailService,
		Data: map[string]interface{}{
			"email":           input.Email,
			"activation_code": token,
			"template_name":   "forgot_password.html",
			"userName":        userName,
		},
	}

	err = ctrl.rabbitMQ.PublishMessage(context.Background(), userActiveMessage)
	if err != nil {
		log.Printf("Kullanıcı oluşturma mesajı gönderilemedi: %v", err)

	}
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"message": "Password reset token sent",
		// "token":   token,
	})
}

func (ctrl *AuthController) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var input dto.ResetPasswordDto
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Geçersiz veri")
		return
	}
	token, err := ctrl.authService.ResetPassword(&input)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"message": token,
	})
}
