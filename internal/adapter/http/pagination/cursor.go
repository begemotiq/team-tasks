package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"task-service/internal/domain"
	"task-service/internal/domain/models"
)

type taskCursorDTO struct {
	CreatedAt time.Time `json:"created_at"`
	ID        int64     `json:"id"`
}

func EncodeTaskCursor(cursor *models.TaskCursor) string {
	if cursor == nil || !cursor.Valid() {
		return ""
	}
	raw, _ := json.Marshal(taskCursorDTO{
		CreatedAt: cursor.CreatedAt,
		ID:        cursor.ID,
	})
	return base64.RawURLEncoding.EncodeToString(raw)
}

func DecodeTaskCursor(value string) (*models.TaskCursor, error) {
	if value == "" {
		return nil, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid cursor", domain.ErrInvalidInput)
	}
	var dto taskCursorDTO
	if err := json.Unmarshal(raw, &dto); err != nil {
		return nil, fmt.Errorf("%w: invalid cursor", domain.ErrInvalidInput)
	}
	cursor := models.TaskCursor{
		CreatedAt: dto.CreatedAt,
		ID:        dto.ID,
	}
	if !cursor.Valid() {
		return nil, fmt.Errorf("%w: invalid cursor", domain.ErrInvalidInput)
	}
	return &cursor, nil
}
