package app

import (
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/huguescodeur/insta-lite/internal/auth"
	"github.com/huguescodeur/insta-lite/internal/middlewares"
	"github.com/huguescodeur/insta-lite/internal/upload"
	httpSwagger "github.com/swaggo/http-swagger"
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
			apiGlobalLimiter := middlewares.ConfigurableRateLimiter("api_global", 10, time.Second)

			r.Group(func(r chi.Router) {
				r.Use(apiGlobalLimiter)
				r.Mount("/auth", a.AuthHandler.AuthRoutes())
			})

			r.Group(func(r chi.Router) {
				jwtSecret := os.Getenv("JWT_SECRET")
				if jwtSecret == "" {
					jwtSecret = "mon_secret"
				}
				authMW := auth.NewAuthMiddleware(a.RedisBlacklister, jwtSecret)
				r.Use(authMW.Handler)

				r.Post("/auth/logout", a.AuthHandler.LogoutHandler)

				r.Route("/upload", func(r chi.Router) {
					r.Get("/signed-url", upload.HandleSignedURL)
				})

				r.Get("/users/{userID}", a.UserHandler.GetByIDHandler)
				r.Mount("/notifications", a.NotificationHandler.NotificationRoutes())
				r.Mount("/feed", a.FeedHandler.FeedRoutes())
				r.Mount("/posts", a.PostHandler.PostRoutes())
				r.Mount("/posts/{postID}/likes", a.LikeHandler.LikeRoutes())

				r.Mount("/posts/{postID}/comments", a.CommentHandler.CommentRoutes())

				r.Route("/comments/{id}", func(r chi.Router) {
					r.Get("/replies", a.CommentHandler.GetRepliesHandler)
					r.Patch("/", a.CommentHandler.UpdateHandler)
					r.Delete("/", a.CommentHandler.DeleteHandler)
				})

				r.Mount("/comments/{commentID}/likes", a.CommentLikeHandler.CommentLikeRoutes())

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
