package routes

import (
	"net/http"

	"github.com/MKMuhammetKaradag/go-microservice/shared/middlewares"
	"github.com/MKMuhammetKaradag/go-microservice/user-service/controllers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func CreateServer() *chi.Mux {
	userController := controllers.NewUserController()
	r := chi.NewRouter()

	r.Route("/auth", func(r chi.Router) {
		r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			render.JSON(w, r, map[string]string{
				"message": "user get",
			})
		})

		r.Group(func(protectedRouter chi.Router) {
			protectedRouter.Use(middlewares.AuthMiddleware)
			protectedRouter.Post("/user", userController.User)
		})
	})
	return r
}
