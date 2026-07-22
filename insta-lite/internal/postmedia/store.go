package postmedia

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type PostMediaStore interface {
	GetByPostIDs(ctx context.Context, postIDs []uuid.UUID) ([]*PostMedia, error)
	Create(ctx context.Context, tx *sql.Tx, media *PostMedia) (*PostMedia, error)
}

type store struct {
	db *sql.DB
}

func NewPostMediaStore(db *sql.DB) PostMediaStore {
	return &store{db: db}
}

func (s *store) GetByPostIDs(ctx context.Context, postIDs []uuid.UUID) ([]*PostMedia, error) {
	q := `
		SELECT media_id, post_id, url, type, created_at 
		FROM post_media
		WHERE post_id = ANY($1)
		`

	raws, err := s.db.QueryContext(ctx, q, postIDs)
	if err != nil {
		return nil, err
	}
	defer raws.Close()

	var postsMedia []*PostMedia

	for raws.Next() {
		pm := PostMedia{}

		if err := raws.Scan(&pm.MediaID, &pm.PostID, &pm.URL, &pm.Type, &pm.CreatedAt); err != nil {
			return nil, err
		}

		postsMedia = append(postsMedia, &pm)
	}

	return postsMedia, nil
}

func (s *store) Create(ctx context.Context, tx *sql.Tx, media *PostMedia) (*PostMedia, error) {
	q := `
		INSERT INTO post_media (media_id, post_id, url, type) 
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
		`

	if err := tx.QueryRowContext(ctx, q, media.MediaID, media.PostID, media.URL, media.Type).Scan(&media.CreatedAt); err != nil {
		return nil, err
	}

	return media, nil
}
