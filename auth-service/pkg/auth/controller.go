package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/MKMuhammetKaradag/go-microservice/auth-service/database"
	"github.com/MKMuhammetKaradag/go-microservice/auth-service/models"
	authMiddleware "github.com/MKMuhammetKaradag/go-microservice/shared/middlewares"
	"github.com/go-chi/render"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func respondWithError(w http.ResponseWriter, code int, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(map[string]string{"error": message})
}



func Register(w http.ResponseWriter, r *http.Request) {
	var user models.User

    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        respondWithError(w, http.StatusBadRequest, "Geçersiz veri")
        return
    }

    // MongoDB bağlantısı
    collection := database.GetCollection("authDB", "users")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Email kontrolü
    emailCount, err := collection.CountDocuments(ctx, bson.M{"email": user.Email})
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Veritabanı hatası")
        return
    }
    if emailCount > 0 {
        respondWithError(w, http.StatusConflict, "Bu email adresi zaten kullanımda")
        return
    }

    // Username kontrolü
    usernameCount, err := collection.CountDocuments(ctx, bson.M{"username": user.Username})
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Veritabanı hatası")
        return
    }
    if usernameCount > 0 {
        respondWithError(w, http.StatusConflict, "Bu kullanıcı adı zaten kullanımda")
        return
    }

    // Şifreyi hashle
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Şifre işlenirken hata oluştu")
        return
    }
    user.Password = string(hashedPassword)

    // Kullanıcıyı veritabanına ekle
    _, err = collection.InsertOne(ctx, user)
    if err != nil {
        // MongoDB duplicate key error kontrolü
        if mongo.IsDuplicateKeyError(err) {
            respondWithError(w, http.StatusConflict, "Bu email veya kullanıcı adı zaten kullanımda")
            return
        }
        respondWithError(w, http.StatusInternalServerError, "Kullanıcı kaydedilemedi")
        return
    }

	w.WriteHeader(http.StatusCreated)
	render.JSON(w,r,map[string]string{
		"message": "Kullanıcı başarıyla oluşturuldu-asa",
	})

	// c.JSON(http.StatusCreated, gin.H{"message": "Kullanıcı başarıyla oluşturuldu"})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var input models.User
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        respondWithError(w, http.StatusBadRequest, "Geçersiz veri")
        return
    }

	collection := database.GetCollection("authDB","users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		respondWithError(w,http.StatusUnauthorized, "Geçersiz e-posta")
		return
	}

	// Şifreyi doğrula
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		respondWithError(w,http.StatusUnauthorized, "yanlış şifre")
		return
	}

	// Redis'e oturum kaydet
	sessionKey := "session:" + user.ID
	fmt.Println(sessionKey)


	userData := map[string]string{
		"email":   user.Email,
		"username": user.Username,
	}
	userDataJson, err := json.Marshal(userData)
	if err != nil {
		respondWithError(w,http.StatusInternalServerError, "Kullanıcı verisi serileştirilemedi")
		return
	}
	err = database.RedisClient.Set( sessionKey, userDataJson, 24 * time.Hour).Err()
	if err != nil {
		fmt.Println(err)
		respondWithError(w,http.StatusInternalServerError, "Oturum kaydedilemedi")
		return
	}


	cookie := &http.Cookie{
        Name:     "session_id",
        Value:    user.ID,
        Path:     "/",
        MaxAge:   60 * 60*24,  // 30 dakika
        HttpOnly: true,
        Secure:   false,     // HTTPS kullanıyorsanız true yapın
        SameSite: http.SameSiteLaxMode,
    }
    http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
	render.JSON(w,r,map[string]string{
		"message": "Giriş başarılı",
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	
	cookieSessionId, err := r.Cookie("session_id")
	
	if err != nil {
		// Return unauthorized if no session token exists
		respondWithError(w,http.StatusInternalServerError, "Giriş yapılmamış")
		// c.JSON(http.StatusUnauthorized, gin.H{"error": "Giriş yapılmamış"})
		// c.Abort()
		return
	}

	sessionKey := "session:" + cookieSessionId.Value
	err = database.RedisClient.Del(sessionKey).Err()
	if err != nil {
		respondWithError(w,http.StatusInternalServerError, "Oturum sonlandırılamadı")
		// c.JSON(http.StatusInternalServerError, gin.H{"error": "Oturum sonlandırılamadı"})
		return
	}
	cookie := &http.Cookie{
        Name:     "session_id",
        Value:    "",
        Path:     "/",
        MaxAge:   -1,        // Cookie'yi hemen sil
        HttpOnly: true,
    }
    http.SetCookie(w, cookie)


	
	w.WriteHeader(http.StatusOK)
	render.JSON(w,r,map[string]string{
		"message": "Başarıyla çıkış yapıldı",
	})
}

func Protected(	w http.ResponseWriter, r *http.Request){

	userData, ok := authMiddleware.GetUserData(r)
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
