package routes

import (
	"github.com/MKMuhammetKaradag/go-microservice/auth-service/pkg/auth"
	authMiddleware "github.com/MKMuhammetKaradag/go-microservice/shared/middlewares"
	"github.com/go-chi/chi/v5"
)

func CreateServer() *chi.Mux{
	r := chi.NewRouter()
	// r.Use(authMiddleware.AuthMiddleware)
	r.Post("/register", auth.Register)
	r.Post("/login", auth.Login)
	// r.Post("/logout", auth.Logout)
	

	protectedRouter := chi.NewRouter()
    protectedRouter.Use(authMiddleware.AuthMiddleware)
    protectedRouter.Post("/logout", auth.Logout)
    protectedRouter.Get("/protected", auth.Protected)
    
    r.Mount("/", protectedRouter)
	return r
}




// func Login(c *gin.Context) {
// 	var input models.User
// 	if err := c.ShouldBindJSON(&input); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz giriş"})
// 		return
// 	}

// 	collection := database.GetCollection("authDB","users")
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	var user models.User
// 	err := collection.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user)
// 	if err != nil {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Geçersiz e-posta veya şifre"})
// 		return
// 	}

// 	// Şifreyi doğrula
// 	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
// 	if err != nil {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Geçersiz e-posta veya şifre"})
// 		return
// 	}

// 	// Redis'e oturum kaydet
// 	sessionKey := "session:" + user.ID
// 	fmt.Println(sessionKey)


// 	userData := map[string]string{
// 		"email":   user.Email,
// 		"username": user.Username,
// 	}
// 	userDataJson, err := json.Marshal(userData)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kullanıcı verisi serileştirilemedi"})
// 		return
// 	}
// 	err = database.RedisClient.Set( sessionKey, userDataJson, 24 * time.Hour).Err()
// 	if err != nil {
// 		fmt.Println(err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Oturum kaydedilemedi"})
// 		return
// 	}

// 	c.SetCookie("session_id", user.ID, 30*60, "/", "", false, true)
// 	c.JSON(http.StatusOK, gin.H{"message": "Giriş başarılı"})
// }

// func Logout(c *gin.Context) {
// 	tokenString, err := c.Cookie("session_id")
	
// 	if err != nil {
// 		// Return unauthorized if no session token exists
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Giriş yapılmamış"})
// 		c.Abort()
// 		return
// 	}

// 	sessionKey := "session:" + tokenString
// 	err = database.RedisClient.Del(sessionKey).Err()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Oturum sonlandırılamadı"})
// 		return
// 	}
// 	c.SetCookie("session_id", "", -1, "/", "", false, true)


// 	c.JSON(http.StatusOK, gin.H{"message": "Başarıyla çıkış yapıldı"})
// }