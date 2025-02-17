package middlewares

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/MKMuhammetKaradag/go-microservice/shared/redisrepo"
)

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

type AuthMiddleware struct {
	redisRepo *redisrepo.RedisRepository
}

func NewAuthMiddleware(redisRepo *redisrepo.RedisRepository) *AuthMiddleware {
	return &AuthMiddleware{redisRepo: redisRepo}
}

// AuthMiddleware is the JWT validation middleware
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) { // Try to fetch the session token from the cookie
			publicRoutes := map[string]bool{
				"/register": true,
				"/login":    true,
			}

			if publicRoutes[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}
			var token string

			// WebSocket isteği mi?
			if strings.Contains(r.Header.Get("Connection"), "Upgrade") && r.Header.Get("Upgrade") == "websocket" {
				// Token’ı URL parametresinden veya `Sec-WebSocket-Protocol` başlığından al
				token = r.URL.Query().Get("token")
				if token == "" {
					token = r.Header.Get("session_id")
					fmt.Println("geldi", token)
				}
			} else {
				// Normal HTTP istekleri için `session_id` çerezini kontrol et
				cookieSessionId, err := r.Cookie("session_id")
				if err != nil {
					respondWithError(w, http.StatusUnauthorized, "Unauthorized: missing session")
					return
				}
				token = "session:" + cookieSessionId.Value
			}

			// Construct the session key for Redis
			// sessionKey := "session:" + cookieSessionId.Value

			// Try to fetch session data from Redis

			userData, err := m.redisRepo.GetSession(token)
			// database.RedisClient.Get(sessionKey).Result()
			if err != nil {
				// Handle case where session data is not found in Redis
				respondWithError(w, http.StatusUnauthorized, "geçersiz oturum")
				return
			}

			// var userData map[string]string
			// err = json.Unmarshal([]byte(tokenRedis), &userData)
			// if err != nil {
			// 	respondWithError(w, http.StatusInternalServerError, "Veri çözümleme hatası")
			// 	return
			// }

			ctx := context.WithValue(r.Context(), "userData", userData)
			next.ServeHTTP(w, r.WithContext(ctx))

		})
}

func GetUserData(r *http.Request) (map[string]string, bool) {
	userData, ok := r.Context().Value("userData").(map[string]string)
	return userData, ok
}
