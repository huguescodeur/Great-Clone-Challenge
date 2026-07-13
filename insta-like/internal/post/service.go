package post

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/huguescodeur/insta-like/internal/postmedia"
	"github.com/huguescodeur/insta-like/internal/user"
)

type PostService struct {
	postStore      PostStore
	postMediaStore postmedia.PostMediaStore
	userStore      user.UserStore
	db             *sql.DB
}

func NewServicePost(p PostStore, m postmedia.PostMediaStore, u user.UserStore, db *sql.DB) *PostService {
	return &PostService{postStore: p, postMediaStore: m, userStore: u, db: db}
}

func (s *PostService) GetAllWithMedia(ctx context.Context, cursorStr string, limit int) (*PaginatedPostResponse, error) {
	cursor, err := DecodeCursor(cursorStr)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor: %w", err)
	}

	posts, err := s.postStore.GetAll(ctx, cursor, limit+1)
	if err != nil {
		return nil, err
	}

	hasMore := len(posts) > limit
	if hasMore {
		posts = posts[:limit]
	}

	var ids []uuid.UUID
	for _, p := range posts {
		ids = append(ids, p.PostID)
	}
	media, err := s.postMediaStore.GetByPostIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	mediaMap := make(map[uuid.UUID][]postmedia.PostMedia)
	for _, m := range media {
		mediaMap[m.PostID] = append(mediaMap[m.PostID], *m)
	}
	var postResponses []*PostResponse
	for _, p := range posts {
		postResponses = append(postResponses, &PostResponse{
			Post:   *p,
			Medias: mediaMap[p.PostID],
		})
	}

	nextCursor := ""
	if hasMore && len(posts) > 0 {
		lastPost := posts[len(posts)-1]
		nextCursor = EncodeCursor(lastPost.CreatedAt, lastPost.PostID)
	}

	return &PaginatedPostResponse{
		Data:       postResponses,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (s *PostService) GetAllWithMediaByID(ctx context.Context, userID uuid.UUID, cursorStr string, limit int) (*PaginatedPostResponse, error) {
	cursor, err := DecodeCursor(cursorStr)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor: %w", err)
	}

	posts, err := s.postStore.GetAllByID(ctx, userID, cursor, limit+1)
	if err != nil {
		return nil, err
	}

	hasMore := len(posts) > limit
	if hasMore {
		posts = posts[:limit]
	}

	var ids []uuid.UUID
	for _, p := range posts {
		ids = append(ids, p.PostID)
	}

	media, err := s.postMediaStore.GetByPostIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	mediaMap := make(map[uuid.UUID][]postmedia.PostMedia)
	for _, m := range media {
		mediaMap[m.PostID] = append(mediaMap[m.PostID], *m)
	}

	var postResponses []*PostResponse
	for _, p := range posts {
		postResponses = append(postResponses, &PostResponse{
			Post:   *p,
			Medias: mediaMap[p.PostID],
		})
	}

	nextCursor := ""
	if hasMore && len(posts) > 0 {
		lastPost := posts[len(posts)-1]
		nextCursor = EncodeCursor(lastPost.CreatedAt, lastPost.PostID)
	}

	return &PaginatedPostResponse{
		Data:       postResponses,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (s *PostService) UpdatePost(ctx context.Context, content string, userID, postID uuid.UUID) (*Post, error) {
	p, err := s.postStore.Update(ctx, content, userID, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil,
				errors.New("post introuvable ou vous n'avez pas l'autorisation de le modifier")
		}

		return nil, err
	}

	return p, nil
}

func (s *PostService) DeletePost(ctx context.Context, userID, postID uuid.UUID) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.postStore.Delete(ctx, tx, userID, postID); err != nil {
		return err
	}

	if _, err := s.userStore.UpdatePostCount(ctx, tx, userID, -1); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *PostService) CreatePost(ctx context.Context, post Post, medias []postmedia.PostMedia) (*PostResponse, error) {
	if len(medias) == 0 {
		return nil, errors.New("un post doit contenir au moins un média")
	}

	post.Content = strings.TrimSpace(post.Content)

	if post.Content == "" {
		return nil, errors.New("la description ne peut être vide")
	}

	if utf8.RuneCountInString(post.Content) > 1000 {
		return nil, errors.New("description trop longue max 1000 caractères")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	post.PostID = uuid.New()

	insertedPost, err := s.postStore.Create(ctx, tx, &post)
	if err != nil {
		return nil, err
	}

	var insertedMedia []postmedia.PostMedia
	for _, media := range medias {
		media.MediaID = uuid.New()
		media.PostID = insertedPost.PostID
		m, err := s.postMediaStore.Create(ctx, tx, &media)
		if err != nil {
			return nil, err
		}

		insertedMedia = append(insertedMedia, *m)
	}

	if _, err := s.userStore.UpdatePostCount(ctx, tx, insertedPost.PostID, 1); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &PostResponse{
		Post:   *insertedPost,
		Medias: insertedMedia,
	}, nil
}
