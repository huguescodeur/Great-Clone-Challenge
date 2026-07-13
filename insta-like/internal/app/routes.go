package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/huguescodeur/insta-like/internal/auth"
	"github.com/huguescodeur/insta-like/internal/middlewares"
)

func (a *App) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middlewares.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173", "http://127.0.0.1:5173", "http://localhost:5174", "http://127.0.0.1:5174"},

		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},

		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "ngrok-skip-browser-warning"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Route("/api", func(r chi.Router) {

		r.Route("/v1", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(middlewares.RateLimiter)
				r.Mount("/auth", a.AuthHandler.AuthRoutes())
			})

			r.Group(func(r chi.Router) {
				r.Use(auth.AuthMidlleware)

				// r.Mount("/users", a.UserHandler.UserRoutes())
				r.Mount("/posts", a.PostHandler.PostRoutes())
				r.Mount("/posts/{postID}/likes", a.LikeHandler.LikeRoutes())

				// r.Route("/posts/{postID}/likes", func(r chi.Router) {
				// 	r.Get("/", a.LikeHandler.GetAllHandler)
				// 	r.Get("/me", a.LikeHandler.GetByUserAndPostHandler)
				// 	r.Patch("/", a.LikeHandler.UpdateReactionHandler)
				// 	r.Delete("/", a.LikeHandler.DeleteHandler)
				// 	r.Post("/", a.LikeHandler.CreateHandler)
				// })

				r.Mount("/posts/{postID}/comments", a.CommentHandler.CommentRoutes())
				// r.Route("/posts/{postID}/comments", func(r chi.Router) {
				// 	r.Get("/", a.CommentHandler.GetAllByPostIDHandler)
				// 	r.Post("/", a.CommentHandler.CreateHandler)
				// })

				r.Route("/comments/{id}", func(r chi.Router) {
					r.Get("/replies", a.CommentHandler.GetRepliesHandler)
					r.Patch("/", a.CommentHandler.UpdateHandler)
					r.Delete("/", a.CommentHandler.DeleteHandler)
				})

				r.Mount("/comments/{commentID}/likes", a.CommentLikeHandler.CommentLikeRoutes())
				// r.Route("/comments/{commentID}/likes", func(r chi.Router) {
				// 	r.Get("/", a.CommentLikeHandler.GetAllHandler)
				// 	r.Get("/me", a.CommentLikeHandler.GetByUserAndCommentHandler)
				// 	r.Patch("/", a.CommentLikeHandler.UpdateReactionHandler)
				// 	r.Delete("/", a.CommentLikeHandler.DeleteHandler)
				// 	r.Post("/", a.CommentLikeHandler.CreateHandler)
				// })

				r.Route("/users/{userID}/follow", func(r chi.Router) {
					r.Post("/", a.FollowHandler.CreateHandler)
					r.Delete("/", a.FollowHandler.DeleteHandler)
					r.Get("/status", a.FollowHandler.GetStatusHandler)
				})
				r.Get("/users/{userID}/followers", a.FollowHandler.GetFollowersHandler)
				r.Get("/users/{userID}/following", a.FollowHandler.GetFollowingHandler)

			})

		})

	})

	return r
}
