package file

import (
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
)

type FileDB struct {
	ID        uuid.UUID
	URL       string
	Filename  string
	Type      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func FileDBToModel(db *FileDB) *models.File {
	return &models.File{
		ID:        db.ID,
		URL:       db.URL,
		Filename:  db.Filename,
		Type:      db.Type,
		CreatedAt: db.CreatedAt,
		UpdatedAt: db.UpdatedAt,
	}
}
