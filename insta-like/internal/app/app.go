package app

import (
	"database/sql"

	"github.com/huguescodeur/insta-like/internal/auth"
	"github.com/huguescodeur/insta-like/internal/comment"
	"github.com/huguescodeur/insta-like/internal/commentlike"
	"github.com/huguescodeur/insta-like/internal/follow"
	"github.com/huguescodeur/insta-like/internal/like"
	"github.com/huguescodeur/insta-like/internal/post"
	"github.com/huguescodeur/insta-like/internal/postmedia"
	"github.com/huguescodeur/insta-like/internal/user"
)

type App struct {
	AuthHandler        *auth.AuthHandler
	PostHandler        *post.PostHandler
	LikeHandler        *like.LikeHandler
	CommentHandler     *comment.CommentHandler
	CommentLikeHandler *commentlike.CommentLikeHandler
	FollowHandler      *follow.FollowHandler
}

func Init(db *sql.DB) *App {
	// ? Store
	authStore := auth.NewAuthStore(db)
	userStore := user.NewUserStore(db)
	postStore := post.NewPostStore(db)
	postMediaStore := postmedia.NewPostMediaStore(db)
	likeStore := like.NewLikeStore(db)
	commentStore := comment.NewCommentStore(db)
	commentLikeStore := commentlike.NewCommentLikeStore(db)
	followStore := follow.NewFollowStore(db)

	// ? Service
	authService := auth.NewServiceAuth(authStore)
	postService := post.NewServicePost(postStore, postMediaStore, userStore, db)
	likeService := like.NewServiceLike(likeStore, postStore, db)
	commentService := comment.NewServiceComment(commentStore, postStore, db)
	commentLikeService := commentlike.NewServiceCommentLike(commentLikeStore, commentStore, db)
	followService := follow.NewServiceFollow(followStore, userStore, db)

	// ? Handler
	authHandler := auth.NewHandlerAuth(authService)
	postHandler := post.NewPostHandler(postService)
	likeHandler := like.NewLikeHandler(likeService)
	commentHandler := comment.NewCommentHandler(commentService)
	commentLikeHandler := commentlike.NewCommentLikeHandler(commentLikeService)
	followHandler := follow.NewFollowHandler(followService)

	return &App{
		AuthHandler:        authHandler,
		PostHandler:        postHandler,
		LikeHandler:        likeHandler,
		CommentHandler:     commentHandler,
		CommentLikeHandler: commentLikeHandler,
		FollowHandler:      followHandler,
	}
}
