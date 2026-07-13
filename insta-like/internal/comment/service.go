package comment

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type PostUpdater interface {
	UpdateCommentsCount(ctx context.Context, tx *sql.Tx, postID uuid.UUID, delta int) error
}

type CommentService struct {
	commentStore CommentStore
	postUpdater  PostUpdater
	db           *sql.DB
}

func NewServiceComment(commentStore CommentStore, postUpdater PostUpdater, db *sql.DB) *CommentService {
	return &CommentService{commentStore: commentStore, postUpdater: postUpdater, db: db}
}

func (s *CommentService) GetAllByPostID(ctx context.Context, postID uuid.UUID, cursorStr string, limit int) (*PaginatedCommentResponse, error) {
	cursor, err := DecodeCursor(cursorStr)
	if err != nil {
		return nil, err
	}

	comments, err := s.commentStore.GetAllByPostID(ctx, postID, cursor, limit+1)
	if err != nil {
		return nil, err
	}

	hasMore := len(comments) > limit
	if hasMore {
		comments = comments[:limit]
	}

	nextCursor := ""
	if hasMore && len(comments) > 0 {
		last := comments[len(comments)-1]
		nextCursor = EncodeCursor(last.CreatedAt, last.CommentID)
	}

	return &PaginatedCommentResponse{
		Data:       comments,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (s *CommentService) GetReplies(ctx context.Context, parentCommentID uuid.UUID) ([]*Comment, error) {
	return s.commentStore.GetRepliesByParentID(ctx, parentCommentID)
}

func (s *CommentService) UpdateComment(ctx context.Context, content string, userID, commentID uuid.UUID) (*Comment, error) {
	c, err := s.commentStore.Update(ctx, content, userID, commentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}

	return c, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, userID, commentID uuid.UUID) error {
	c, err := s.commentStore.GetByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCommentNotFound
		}
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.commentStore.Delete(ctx, tx, userID, commentID); err != nil {
		return err
	}

	if err := s.postUpdater.UpdateCommentsCount(ctx, tx, c.PostID, -1); err != nil {
		return ErrUpdateCountFailed
	}

	return tx.Commit()
}

func (s *CommentService) CreateComment(ctx context.Context, comment Comment) (*Comment, error) {
	comment.Content = strings.TrimSpace(comment.Content)

	if comment.Content == "" {
		return nil, errors.New("le commentaire ne peut être vide")
	}
	if utf8.RuneCountInString(comment.Content) > 2000 {
		return nil, errors.New("commentaire trop long, max 2000 caractères")
	}

	if comment.ParentCommentID != nil {
		parent, err := s.commentStore.GetByID(ctx, *comment.ParentCommentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrParentCommentNotFound
			}
			return nil, err
		}
		if parent.PostID != comment.PostID {
			return nil, ErrCommentPostMismatch
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	comment.CommentID = uuid.New()

	inserted, err := s.commentStore.Create(ctx, tx, &comment)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, ErrPostNotFound
		}
		return nil, ErrCreateCommentFailed
	}

	if err := s.postUpdater.UpdateCommentsCount(ctx, tx, inserted.PostID, 1); err != nil {
		return nil, ErrUpdateCountFailed
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return inserted, nil
}
