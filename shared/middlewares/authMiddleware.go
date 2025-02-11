package authMiddleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/MKMuhammetKaradag/go-microservice/shared/database"
)

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// AuthMiddleware is the JWT validation middleware
func AuthMiddleware(next http.Handler) http.Handler {
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
			cookieSessionId, err := r.Cookie("session_id")
			if err != nil {

				respondWithError(w, http.StatusUnauthorized, err.Error())
				return
			}

			// Construct the session key for Redis
			sessionKey := "session:" + cookieSessionId.Value

			// Try to fetch session data from Redis
			tokenRedis, err := database.RedisClient.Get(sessionKey).Result()
			if err != nil {
				// Handle case where session data is not found in Redis
				respondWithError(w, http.StatusUnauthorized, "geçersiz oturum")
				return
			}

			var userData map[string]string
			err = json.Unmarshal([]byte(tokenRedis), &userData)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Veri çözümleme hatası")
				return
			}

			ctx := context.WithValue(r.Context(), "userData", userData)
			next.ServeHTTP(w, r.WithContext(ctx))

		})
}

func GetUserData(r *http.Request) (map[string]string, bool) {
	userData, ok := r.Context().Value("userData").(map[string]string)
	return userData, ok
}
