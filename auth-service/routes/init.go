package routes

import (
	"github.com/MKMuhammetKaradag/go-microservice/auth-service/controllers"

	"github.com/MKMuhammetKaradag/go-microservice/auth-service/repository"
	"github.com/MKMuhammetKaradag/go-microservice/auth-service/websocket"
	"github.com/MKMuhammetKaradag/go-microservice/shared/messaging"
	"github.com/MKMuhammetKaradag/go-microservice/shared/middlewares"
	"github.com/MKMuhammetKaradag/go-microservice/shared/redisrepo"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

// CreateServer: Router oluşturur ve tüm endpointleri ekler
func CreateServer(rabbitMQ *messaging.RabbitMQ, sessionRepo *redisrepo.RedisRepository, userRepo *repository.UserRepository) *chi.Mux {
	authController := controllers.NewAuthController(rabbitMQ, sessionRepo)
	authMiddleware := middlewares.NewAuthMiddleware(sessionRepo)
	hub := websocket.NewHub()

	go hub.Run()
	go hub.ListenRedisStatus(sessionRepo)

	wsController := controllers.NewWebSocketController(hub, userRepo, sessionRepo)
	r := chi.NewRouter()

	// Global Middleware'ler
	r.Use(middlewares.PrometheusMiddleware)

	// Servis Route'larını Gruplama
	registerMetricsRoutes(r)
	registerAuthRoutes(r, authController, authMiddleware, wsController)
	registerSwaggerRoutes(r)

	return r
}

// Prometheus ve metrik endpointleri ekler
func registerMetricsRoutes(r *chi.Mux) {
	r.Mount("/metrics", promhttp.Handler())
}

// Auth ile ilgili tüm endpointleri ekler
func registerAuthRoutes(r *chi.Mux, authController *controllers.AuthController, authMiddleware *middlewares.AuthMiddleware, wsController *controllers.WebSocketController) {
	r.Route("/auth", func(r chi.Router) {
		r.Use(middlewares.Logger) // Tüm /auth endpointlerinde logger middleware aktif olacak

		// Public endpointler
		r.Post("/signUp", authController.SignUp)
		r.Post("/activationUser", authController.ActivationUser)
		r.Post("/signIn", authController.SignIn)
		r.Post("/forgotPassword", authController.ForgotPassword)
		r.Post("/resetPassword", authController.ResetPassword)

		// Protected Routes (JWT Authentication Gerekli)
		r.Group(func(protectedRouter chi.Router) {
			protectedRouter.Use(authMiddleware.Authenticate)
			protectedRouter.Post("/logout", authController.Logout)
			protectedRouter.Get("/me", authController.Logout)
			protectedRouter.Post("/updateStatus", authController.UpdateStatus)
			protectedRouter.Get("/ws", wsController.HandleWebSocket)
		})
	})
}

// Swagger dökümantasyonunu ekler
func registerSwaggerRoutes(r *chi.Mux) {
	r.Get("/swagger/*", httpSwagger.WrapHandler)
}
