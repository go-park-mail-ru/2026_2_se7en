package session

import "github.com/go-park-mail-ru/2026_2_se7en/internal/models"

func SessionDBToModel(db *SessionDB) *models.Session {
	if db == nil {
		return nil
	}
	return &models.Session{
		ID:        db.ID,
		UserID:    db.UserID,
		ExpiresAt: db.ExpiresAt,
		CreatedAt: db.CreatedAt,
		UpdatedAt: db.UpdatedAt,
	}
}
