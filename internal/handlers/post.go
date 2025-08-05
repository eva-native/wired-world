package handlers

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/eva-native/wired-world/internal/data"
	"github.com/eva-native/wired-world/internal/repository"
	"github.com/eva-native/wired-world/web"
)

const (
	messageFormName = "message"
)

var tmpl = template.Must(template.ParseFS(web.Templates, "*/**"))

func AllPost(posts repository.Posts) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
		defer cancel()

		ps, err := posts.All(ctx)
		if err != nil {
			log.Printf("[%s]: get all posts: %s", r.RemoteAddr, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if err := tmpl.ExecuteTemplate(w, "posts.html", ps); err != nil {
			log.Printf("[%s]: render all posts: %s", r.RemoteAddr, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func AddNewPost(posts repository.Posts) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
		defer cancel()
		if err := r.ParseForm(); err != nil {
			log.Printf("[%s]: %s", r.RemoteAddr, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		m := data.PrepareMessage(r.FormValue(messageFormName))
		if err := data.ValidateMessage(m); err != nil {
			log.Printf("[%s]: %s", r.RemoteAddr, err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		p, err := posts.Add(ctx, time.Now(), m)
		if err != nil {
			log.Printf("[%s]: add post: %s", r.RemoteAddr, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if err := tmpl.ExecuteTemplate(w, "post.html", &p); err != nil {
			log.Printf("[%s]: render added post: %s", r.RemoteAddr, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}
