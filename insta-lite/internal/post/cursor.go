package post

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func EncodeCursor(t time.Time, id uuid.UUID) string {
	if t.IsZero() {
		return ""
	}
	b := fmt.Sprintf("%s,%s", t.Format(time.RFC3339Nano), id.String())
	return base64.StdEncoding.EncodeToString([]byte(b))
}

func DecodeCursor(cursorStr string) (*PostCursor, error) {
	if cursorStr == "" {
		return nil, nil
	}
	b, err := base64.StdEncoding.DecodeString(cursorStr)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(string(b), ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid cursor format")
	}

	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(parts[1])
	if err != nil {
		return nil, err
	}

	return &PostCursor{CreatedAt: t, PostID: id}, nil
}
