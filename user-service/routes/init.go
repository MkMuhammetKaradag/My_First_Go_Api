package routes

import (
	"net/http"

	authMiddleware "github.com/MKMuhammetKaradag/go-microservice/shared/middlewares"
	"github.com/MKMuhammetKaradag/go-microservice/user-service/controllers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func CreateServer() *chi.Mux {
	userController := controllers.NewUserController()
	r := chi.NewRouter()
	// r.Use(authMiddleware.AuthMiddleware)
	r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, map[string]string{
			"message": "userslar çekildi",
		})
	})

	// r.Post("/logout", auth.Logout)

	protectedRouter := chi.NewRouter()
	protectedRouter.Use(authMiddleware.AuthMiddleware)
	protectedRouter.Post("/user", userController.User)
	r.Mount("/", protectedRouter)
	return r
}
