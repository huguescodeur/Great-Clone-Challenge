package notification

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type NotificationStore interface {
	Create(ctx context.Context, n *Notification) (*Notification, error)
	GetAllByRecipient(ctx context.Context, recipientID uuid.UUID, cursor *NotificationCursor, limit int) ([]*Notification, error)
	CountUnread(ctx context.Context, recipientID uuid.UUID) (int64, error)
	MarkAsRead(ctx context.Context, recipientID, notificationID uuid.UUID) error
}

type store struct {
	db *sql.DB
}

func NewNotificationStore(db *sql.DB) NotificationStore {
	return &store{db: db}
}

func (s *store) Create(ctx context.Context, n *Notification) (*Notification, error) {
	q := `
		INSERT INTO notifications (recipient_id, actor_id, type, entity_id)
		VALUES ($1, $2, $3, $4)
		RETURNING notification_id, recipient_id, actor_id, type, entity_id, read, created_at
		`

	if err := s.db.QueryRowContext(ctx, q, n.RecipientID, n.ActorID, n.Type, n.EntityID).Scan(
		&n.NotificationID, &n.RecipientID, &n.ActorID, &n.Type, &n.EntityID, &n.Read, &n.CreatedAt,
	); err != nil {
		return nil, err
	}
	return n, nil
}

func (s *store) GetAllByRecipient(ctx context.Context, recipientID uuid.UUID, cursor *NotificationCursor, limit int) ([]*Notification, error) {
	var q string
	var rows *sql.Rows
	var err error

	if cursor == nil {
		q = `
			SELECT notification_id, recipient_id, actor_id, type, entity_id, read, created_at
			FROM notifications
			WHERE recipient_id = $1
			ORDER BY created_at DESC, notification_id DESC
			LIMIT $2
			`
		rows, err = s.db.QueryContext(ctx, q, recipientID, limit)
	} else {
		q = `
			SELECT notification_id, recipient_id, actor_id, type, entity_id, read, created_at
			FROM notifications
			WHERE recipient_id = $1
			  AND (created_at < $2 OR (created_at = $3 AND notification_id < $4))
			ORDER BY created_at DESC, notification_id DESC
			LIMIT $5
			`
		rows, err = s.db.QueryContext(ctx, q, recipientID, cursor.CreatedAt, cursor.CreatedAt, cursor.NotificationID, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := make([]*Notification, 0)
	for rows.Next() {
		n := Notification{}
		if err := rows.Scan(&n.NotificationID, &n.RecipientID, &n.ActorID, &n.Type, &n.EntityID, &n.Read, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, &n)
	}
	return notifications, rows.Err()
}

func (s *store) CountUnread(ctx context.Context, recipientID uuid.UUID) (int64, error) {
	var count int64
	q := `SELECT COUNT(*) FROM notifications WHERE recipient_id = $1 AND read = false`
	if err := s.db.QueryRowContext(ctx, q, recipientID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *store) MarkAsRead(ctx context.Context, recipientID, notificationID uuid.UUID) error {
	q := `UPDATE notifications SET read = true WHERE notification_id = $1 AND recipient_id = $2`
	result, err := s.db.ExecContext(ctx, q, notificationID, recipientID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotificationNotFound
	}
	return nil
}
