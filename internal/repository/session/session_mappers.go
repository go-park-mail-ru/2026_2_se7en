package session

import "app/internal/models"

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
