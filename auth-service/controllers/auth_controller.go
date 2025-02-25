package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/MKMuhammetKaradag/go-microservice/auth-service/docs"
	"github.com/MKMuhammetKaradag/go-microservice/auth-service/dto"
	"github.com/MKMuhammetKaradag/go-microservice/auth-service/services"
	"github.com/MKMuhammetKaradag/go-microservice/shared/messaging"
	"github.com/MKMuhammetKaradag/go-microservice/shared/middlewares"
	"github.com/MKMuhammetKaradag/go-microservice/shared/models"
	"github.com/MKMuhammetKaradag/go-microservice/shared/redisrepo"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type SwagerSignup struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type SwagerSignin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type SignUpResponse struct {
	Message             string `json:"message"`
	UserActivationToken string `json:"userActivationToken"`
}
type ActivationResponse struct {
	Message string `json:"message"`
	user    string `json:"user"`
}

type LogoutResponse struct {
	Message string `json:"message"`
}

// ErrorResponse Hata mesajlarını döndürmek için kullanılan yapı
type ErrorResponse struct {
	Error string `json:"error"`
}
type AuthController struct {
	authService *services.AuthService
	rabbitMQ    *messaging.RabbitMQ
	sessionRepo *redisrepo.RedisRepository
}

func NewAuthController(rabbitMQ *messaging.RabbitMQ, sessionRepo *redisrepo.RedisRepository) *AuthController {
	return &AuthController{
		authService: services.NewAuthService(),
		rabbitMQ:    rabbitMQ,
		sessionRepo: sessionRepo,
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

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

// @Summary      Kullanıcı Kaydı
// @Description  Yeni bir kullanıcı oluşturur
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body models.User true "Kullanıcı Kayıt Modeli"
// @Success      200  {object}  SignUpResponse
// @Failure      400  {object}  ErrorResponse
// @Router       /auth/signUp [post]
func (ctrl *AuthController) SignUp(w http.ResponseWriter, r *http.Request) {
	var user = models.NewUser()

	// Kullanıcı verisini JSON'dan çözümle
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		respondWithError(w, http.StatusBadRequest, "Geçersiz veri formatı")
		return
	}

	// Kullanıcı verisini doğrula
	if err := validateUser(&user); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Kullanıcıyı kaydet ve aktivasyon bilgilerini al
	activationCode, activationToken, err := ctrl.authService.SignUp(&user)
	if err != nil {
		respondWithError(w, http.StatusConflict, err.Error())
		return
	}

	// Aktivasyon e-postası için mesaj oluştur
	emailMessage := messaging.Message{
		Type:      "active_user",
		ToService: messaging.EmailService,
		Data: map[string]interface{}{
			"email":           user.Email,
			"activation_code": activationCode,
			"template_name":   "activation_email.html",
			"userName":        user.Username,
		},
	}

	// RabbitMQ'ya aktivasyon mesajı gönder
	if err := ctrl.rabbitMQ.PublishMessage(context.Background(), emailMessage); err != nil {
		log.Printf("Kullanıcı aktivasyon mesajı gönderilemedi: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Aktivasyon e-postası gönderilemedi")
		return
	}

	// Başarılı yanıtı dön
	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"message":             "Kullanıcı başarıyla oluşturuldu",
		"userActivationToken": activationToken,
	})
}

// @Summary      Kullanıcı Aktivasyonu
// @Description   kullanıcı etkinleştir
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.ActivationRequest true "Kullanıcı Aktivasyon  Modeli"
// @Success      200  {object}  ActivationResponse
// @Failure      400  {object}  ErrorResponse
// @Router       /auth/activationUser [post]
func (ctrl *AuthController) ActivationUser(w http.ResponseWriter, r *http.Request) {
	var activationRequest dto.ActivationRequest

	// Aktivasyon isteğini JSON'dan çözümle
	if err := json.NewDecoder(r.Body).Decode(&activationRequest); err != nil {
		respondWithError(w, http.StatusBadRequest, "Geçersiz veri formatı")
		return
	}

	// Aktivasyon işlemini gerçekleştir
	activatedUser, err := ctrl.authService.ActivationUser(activationRequest.ActivationCode, activationRequest.ActivationToken)
	if err != nil {
		respondWithError(w, http.StatusConflict, err.Error())
		return
	}

	// Kullanıcı oluşturulduğunda RabbitMQ'ya mesaj gönder
	userCreatedMessage := messaging.Message{
		Type:      "user_created",
		ToService: messaging.UserService, // Eğer belirli bir servis varsa burada belirtin
		Data: map[string]interface{}{
			"user_id":   activatedUser.ID,
			"email":     activatedUser.Email,
			"firstName": activatedUser.FirstName,
			"age":       activatedUser.Age,
			"createdAt": activatedUser.CreatedAt,
			"username":  activatedUser.Username,
		},
	}

	// Mesaj gönderme başarısız olursa hata loglanır ancak işlem devam eder
	if err := ctrl.rabbitMQ.PublishMessage(context.Background(), userCreatedMessage); err != nil {
		log.Printf("Kullanıcı oluşturma mesajı gönderilemedi: %v", err)
	}

	// Başarılı yanıt dön
	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Kullanıcı başarıyla oluşturuldu",
		"user":    activatedUser,
	})
}

// @Summary      Kullanıcı Giriş
// @Description   kullanıcı giriş
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body SwagerSignin true "Kullanıcı giriş  Modeli"
// @Success      200  {object}  ActivationResponse
// @Failure      400  {object}  ErrorResponse
// @Router       /auth/signIn [post]
func (ctrl *AuthController) SignIn(w http.ResponseWriter, r *http.Request) {
	var input models.User
	fmt.Println("auth controller yapısına geldi")

	// Gelen isteği çözümle
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Geçersiz veri formatı")
		return
	}

	// Kullanıcıyı kimlik doğrulama servisine gönder
	user, err := ctrl.authService.SignIn(&input)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Kullanıcı rollerini JSON formatına çevir
	rolesJSON, err := json.Marshal(user.Roles)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Rol bilgisi dönüştürülemedi")
		log.Println("Rol JSON hatası:", err)
		return
	}

	// Redis'te oturum oluştur
	sessionKey := "session:" + user.ID
	userData := map[string]string{
		"id":       user.ID,
		"email":    user.Email,
		"roles":    string(rolesJSON),
		"username": user.Username,
	}

	if err := ctrl.sessionRepo.SetSession(sessionKey, userData, 24*time.Hour); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Oturum kaydedilemedi")
		log.Println("Redis oturum hatası:", err)
		return
	}

	// Kullanıcı için çerez oluştur
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    user.ID,
		Path:     "/",
		MaxAge:   60 * 60 * 24, // 1 gün
		HttpOnly: true,
		Secure:   false, // HTTPS kullanılıyorsa true yapılmalı
		SameSite: http.SameSiteLaxMode,
	})

	// Başarılı yanıt dön
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Giriş başarılı",
		"user":    user,
	})
}

// @Summary      Kullanıcı Çıkışı
// @Description   kullanıcı Çıkışı
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Success      200  {object}   LogoutResponse
// @Failure      400  {object}  ErrorResponse
// @Router       /auth/logout [post]
func (ctrl *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	// Çerezden session_id al
	cookieSessionId, err := r.Cookie("session_id")
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Giriş yapılmamış")
		return
	}

	// Redis'ten oturumu sil
	sessionKey := "session:" + cookieSessionId.Value
	if err := ctrl.sessionRepo.DeleteSession(sessionKey); err != nil {
		log.Println("Redis oturum silme hatası:", err)
	}

	// Çerezi geçersiz hale getir
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Tarayıcıdan hemen silinsin
		HttpOnly: true,
		Secure:   false, // HTTPS kullanılıyorsa true yapılmalı
		SameSite: http.SameSiteStrictMode,
	})

	// Başarılı yanıt döndür
	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Başarıyla çıkış yapıldı",
	})

}

// @Summary      Protected   router
// @Description  otum açmış kullanıcının bilgiyi doner
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Success      200  {object}   ActivationResponse
// @Failure      400  {object}  ErrorResponse
// @Router       /auth/protected [post]
func Protected(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı verisini al
	userData, ok := middlewares.GetUserData(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Yetkisiz erişim")
		return
	}

	// JSON yanıt döndür
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Protected endpoint",
		"user":    userData,
	})
}

// @Summary       şifremi unuttum
// @Description   kullanıcı giriş şifresini unutulduğun email yollar
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.ForgotPasswordDto true "Kullanıcı şifre unutum modeli  Modeli"
// @Success      200  {object}  LogoutResponse
// @Failure      400  {object}  ErrorResponse
// @Router       /auth/forgotPassword [post]
func (ctrl *AuthController) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var input dto.ForgotPasswordDto

	// Gelen JSON'u çözümle
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Geçersiz veri formatı")
		return
	}

	// Şifre sıfırlama tokeni oluştur
	token, userName, err := ctrl.authService.ForgotPassword(input.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "E-posta adresi kayıtlı değil")
		return
	}

	// Şifre sıfırlama e-postası için mesaj hazırla
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

	// RabbitMQ'ya mesaj gönder
	if err := ctrl.rabbitMQ.PublishMessage(context.Background(), userActiveMessage); err != nil {
		log.Printf("Şifre sıfırlama e-postası gönderilemedi: %v", err)
	}

	// Başarı yanıtı döndür
	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Şifre sıfırlama talimatları e-posta adresinize gönderildi",
	})
}

// @Summary       şifremi  değiştirme
// @Description   kullanıcı giriş şifresini yenileme
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.ResetPasswordDto true "Kullanıcı şifre unutum modeli  Modeli"
// @Success      200  {object}  LogoutResponse
// @Failure      400  {object}  ErrorResponse
// @Router       /auth/resetPassword [post]
func (ctrl *AuthController) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var input dto.ResetPasswordDto

	// Gelen JSON'u çözümle
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Geçersiz veri formatı")
		return
	}

	// Şifre sıfırlama işlemini gerçekleştir
	message, err := ctrl.authService.ResetPassword(&input)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Geçersiz ya da süresi dolmuş token")
		return
	}

	// Başarı yanıtı döndür
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": message,
	})
}

func (ctrl *AuthController) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	var input dto.UpdateStatusDto
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Geçersiz veri")
		return
	}
	userData, ok := middlewares.GetUserData(r)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Kullanıcı bilgisi bulunamadı")
		return
	}

	id, exists := userData["id"]
	if !exists {
		fmt.Println("id not found in userData")
	} else {
		fmt.Println("User ID:", id)
	}
	fmt.Println(id, input.Status)
	err := ctrl.authService.UpdateStatus(id, input.Status)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	ctrl.sessionRepo.PublishStatus(id, "away")
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"message": "ok",
	})

}
