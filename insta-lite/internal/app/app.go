package app

import (
	"context"
	"database/sql"
	"log"
	"os"
	"sync"
	"time"

	"github.com/huguescodeur/insta-lite/internal/auth"
	"github.com/huguescodeur/insta-lite/internal/comment"
	"github.com/huguescodeur/insta-lite/internal/commentlike"
	"github.com/huguescodeur/insta-lite/internal/events"
	"github.com/huguescodeur/insta-lite/internal/feed"
	"github.com/huguescodeur/insta-lite/internal/follow"
	"github.com/huguescodeur/insta-lite/internal/like"
	"github.com/huguescodeur/insta-lite/internal/middlewares"
	"github.com/huguescodeur/insta-lite/internal/notification"
	"github.com/huguescodeur/insta-lite/internal/post"
	"github.com/huguescodeur/insta-lite/internal/postmedia"
	"github.com/huguescodeur/insta-lite/internal/user"
	"github.com/redis/go-redis/v9"
)

type App struct {
	AuthHandler         *auth.AuthHandler
	PostHandler         *post.PostHandler
	LikeHandler         *like.LikeHandler
	CommentHandler      *comment.CommentHandler
	CommentLikeHandler  *commentlike.CommentLikeHandler
	FollowHandler       *follow.FollowHandler
	RedisBlacklister    *auth.RedisBlacklist
	FeedHandler         *feed.FeedHandler
	NotificationHandler *notification.NotificationHandler

	eventsCancel context.CancelFunc
	eventsWG     *sync.WaitGroup
}

func Init(db *sql.DB) *App {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	middlewares.InitRateLimiter(rdb)

	// ? Store
	authStore := auth.NewAuthStore(db)
	userStore := user.NewUserStore(db)
	postStore := post.NewPostStore(db)
	postMediaStore := postmedia.NewPostMediaStore(db)
	likeStore := like.NewLikeStore(db)
	commentStore := comment.NewCommentStore(db)
	commentLikeStore := commentlike.NewCommentLikeStore(db)
	followStore := follow.NewFollowStore(db)
	feedStore := feed.NewFeedStore(db)
	notificationStore := notification.NewNotificationStore(db)

	feedCache := feed.NewRedisFeedCache(rdb)
	likeCounter := like.NewRedisLikeCounter(rdb)
	commentsCounter := comment.NewRedisCommentsCounter(rdb)
	commentLikesCounter := commentlike.NewRedisCommentLikesCounter(rdb)
	fanout := feed.NewFanout(followStore, userStore, feedCache)

	// ? Events
	postBus := events.NewTypedBus[events.PostCreatedEvent](100)
	likedBus := events.NewTypedBus[events.PostLikedEvent](100)
	commentBus := events.NewTypedBus[events.CommentCreatedEvent](100)
	commentLikedBus := events.NewTypedBus[events.CommentLikedEvent](100)
	followedBus := events.NewTypedBus[events.UserFollowedEvent](100)
	eventsCtx, eventsCancel := context.WithCancel(context.Background())

	var eventsWG sync.WaitGroup
	events.StartWorkers(eventsCtx, &eventsWG, postBus, 3, fanout.HandlePostCreated)

	// ? Service
	redisBlacklister := auth.NewRedisBlacklist(rdb)
	authService := auth.NewServiceAuth(authStore, redisBlacklister)
	postService := post.NewServicePost(postStore, postMediaStore, userStore, db, postBus, likeCounter, commentsCounter)

	likeService := like.NewServiceLike(likeStore, postStore, likeCounter, postStore, likedBus, db)
	commentService := comment.NewServiceComment(commentStore, postStore, commentsCounter, commentLikesCounter, postStore, commentBus, db)
	commentLikeService := commentlike.NewServiceCommentLike(commentLikeStore, commentStore, commentLikesCounter, commentStore, commentLikedBus, db)
	followService := follow.NewServiceFollow(followStore, userStore, db, followedBus)
	feedService := feed.NewServiceFeed(feedStore, feedCache, followStore, userStore, postStore, postMediaStore, likeCounter, commentsCounter)
	notificationService := notification.NewServiceNotification(notificationStore)

	events.StartWorkers(eventsCtx, &eventsWG, likedBus, 2, notificationService.HandlePostLiked)
	events.StartWorkers(eventsCtx, &eventsWG, commentBus, 2, notificationService.HandleCommentCreated)
	events.StartWorkers(eventsCtx, &eventsWG, commentLikedBus, 2, notificationService.HandleCommentLiked)
	events.StartWorkers(eventsCtx, &eventsWG, followedBus, 2, notificationService.HandleUserFollowed)

	// ? Handler
	authHandler := auth.NewHandlerAuth(authService)
	postHandler := post.NewPostHandler(postService)
	likeHandler := like.NewLikeHandler(likeService)
	commentHandler := comment.NewCommentHandler(commentService)
	commentLikeHandler := commentlike.NewCommentLikeHandler(commentLikeService)
	followHandler := follow.NewFollowHandler(followService)
	feedHandler := feed.NewFeedHandler(feedService)
	notificationHandler := notification.NewNotificationHandler(notificationService)

	return &App{
		AuthHandler:         authHandler,
		PostHandler:         postHandler,
		LikeHandler:         likeHandler,
		CommentHandler:      commentHandler,
		CommentLikeHandler:  commentLikeHandler,
		FollowHandler:       followHandler,
		RedisBlacklister:    redisBlacklister,
		FeedHandler:         feedHandler,
		NotificationHandler: notificationHandler,

		eventsCancel: eventsCancel,
		eventsWG:     &eventsWG,
	}
}

func (a *App) Shutdown() {
	a.eventsCancel()

	done := make(chan struct{})
	go func() {
		a.eventsWG.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("[app] tous les workers ont terminé proprement")
	case <-time.After(10 * time.Second):
		log.Println("[app] timeout shutdown workers, arrêt forcé")
	}
}
