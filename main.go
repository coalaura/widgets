package main

import (
	"net/http"

	"github.com/coalaura/plain"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	manager = NewWidgetManager()
	log     = plain.New(plain.WithDate(plain.RFC3339Local))
)

func main() {
	manager.RegisterDefault()

	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(log.Middleware())

	r.Get("/widgets.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		w.Write(manager.JSON())
	})

	r.Get("/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")

		err := manager.Render(w, r, name)
		if err != nil {
			http.Error(w, "Failed to render widget", http.StatusInternalServerError)
		}
	})

	log.Println("Listening on http://localhost:4777/")
	log.MustFail(http.ListenAndServe(":4777", r))
}
