package routes

import (
	"net/http"

	"github.com/MKMuhammetKaradag/go-microservice/shared/middlewares"
	"github.com/MKMuhammetKaradag/go-microservice/shared/redisrepo"
	"github.com/MKMuhammetKaradag/go-microservice/user-service/controllers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func CreateServer(sessionRepo *redisrepo.RedisRepository) *chi.Mux {
	userController := controllers.NewUserController()
	authMiddleware := middlewares.NewAuthMiddleware(sessionRepo)
	r := chi.NewRouter()

	r.Route("/auth", func(r chi.Router) {
		r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			render.JSON(w, r, map[string]string{
				"message": "user get",
			})
		})

		r.Group(func(protectedRouter chi.Router) {
			protectedRouter.Use(authMiddleware.Authenticate)
			protectedRouter.Post("/user", userController.User)
		})
	})
	return r
}
