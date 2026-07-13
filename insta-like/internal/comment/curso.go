package comment

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func EncodeCursor(createdAt time.Time, commentID uuid.UUID) string {
	raw := fmt.Sprintf("%s|%s", createdAt.Format(time.RFC3339Nano), commentID.String())
	return base64.URLEncoding.EncodeToString([]byte(raw))
}

func DecodeCursor(cursorStr string) (*CommentCursor, error) {
	if cursorStr == "" {
		return nil, nil
	}

	decoded, err := base64.URLEncoding.DecodeString(cursorStr)
	if err != nil {
		return nil, fmt.Errorf("curseur invalide: %w", err)
	}

	parts := strings.SplitN(string(decoded), "|", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("curseur invalide")
	}

	createdAt, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return nil, fmt.Errorf("curseur invalide: %w", err)
	}

	commentID, err := uuid.Parse(parts[1])
	if err != nil {
		return nil, fmt.Errorf("curseur invalide: %w", err)
	}

	return &CommentCursor{CreatedAt: createdAt, CommentID: commentID}, nil
}
