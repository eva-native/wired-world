package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/eva-native/wired-world/internal/data"
	"github.com/eva-native/wired-world/internal/repository"
	"github.com/eva-native/wired-world/web"
)

const (
	messageFormName = "message"
)

var tmpl = web.Templates()

func AllPost(posts repository.Posts, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
		defer cancel()
		ps, err := posts.All(ctx)
		if err != nil {
			logger.Error("get all posts", "remote_addr", r.RemoteAddr, "err", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if err := tmpl.ExecuteTemplate(w, "posts.tmpl", ps); err != nil {
			logger.Error("render all posts", "remote_addr", r.RemoteAddr, "err", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func AddNewPost(posts repository.Posts, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
		defer cancel()
		if err := r.ParseForm(); err != nil {
			logger.Warn("parse form", "remote_addr", r.RemoteAddr, "err", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		m := data.PrepareMessage(r.FormValue(messageFormName))
		if err := data.ValidateMessage(m); err != nil {
			logger.Warn("validate message", "remote_addr", r.RemoteAddr, "err", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if _, err := posts.Add(ctx, time.Now(), m); err != nil {
			logger.Error("add post", "remote_addr", r.RemoteAddr, "err", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ps, err := posts.All(ctx)
		if err != nil {
			logger.Error("get all posts after add", "remote_addr", r.RemoteAddr, "err", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if err := tmpl.ExecuteTemplate(w, "posts.tmpl", ps); err != nil {
			logger.Error("render posts after add", "remote_addr", r.RemoteAddr, "err", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}
