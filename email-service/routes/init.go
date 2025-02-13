package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func CreateServer() *chi.Mux {

	r := chi.NewRouter()

	r.Route("/email", func(r chi.Router) {
		r.Get("/notification", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			render.JSON(w, r, map[string]string{
				"message": "email get notification",
			})
		})

	})
	return r
}
