package routes

import (
	"github.com/MKMuhammetKaradag/go-microservice/auth-service/controllers"
	"github.com/MKMuhammetKaradag/go-microservice/shared/messaging"
	"github.com/MKMuhammetKaradag/go-microservice/shared/middlewares"
	"github.com/MKMuhammetKaradag/go-microservice/shared/redisrepo"
	"github.com/go-chi/chi/v5"
)

func CreateServer(rabbitMQ *messaging.RabbitMQ, sessionRepo *redisrepo.RedisRepository) *chi.Mux {
	authController := controllers.NewAuthController(rabbitMQ, sessionRepo)
	authMiddleware := middlewares.NewAuthMiddleware(sessionRepo)
	r := chi.NewRouter()
	r.Use(middlewares.Logger)
	r.Route("/auth", func(r chi.Router) {
		r.Post("/signUp", authController.SignUp)
		r.Post("/activationUser", authController.ActivationUser)
		r.Post("/signIn", authController.SignIn)
		r.Post("/forgotPassword", authController.ForgotPassword)
		r.Post("/resetPassword", authController.ResetPassword)

		r.Group(func(protectedRouter chi.Router) {
			protectedRouter.Use(authMiddleware.Authenticate)
			protectedRouter.Post("/logout", authController.Logout)
			protectedRouter.Get("/me", authController.Logout)
			protectedRouter.Get("/protected", controllers.Protected)
		})
	})

	return r
}
