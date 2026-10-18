package auth

import "app/models"

func UserDBToModel(db *UserDB) *models.User {
	if db == nil {
		return nil
	}
	return &models.User{
		ID:           db.ID,
		Email:        db.Email,
		PasswordHash: db.PasswordHash,
		PhoneNumber:  db.PhoneNumber,
		CreatedAt:    db.CreatedAt,
		UpdatedAt:    db.UpdatedAt,
		DeletedAt:    db.DeletedAt,
	}
}

func ProfileDBToModel(db *ProfileDB) *models.Profile {
	if db == nil {
		return nil
	}
	return &models.Profile{
		ID:        db.ID,
		IconID:    db.IconID,
		UserID:    db.UserID,
		Nickname:  db.Nickname,
		FirstName: db.FirstName,
		LastName:  db.LastName,
		Bio:       db.Bio,
		CreatedAt: db.CreatedAt,
		UpdatedAt: db.UpdatedAt,
	}
}

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
