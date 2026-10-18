package profile

import "app/models"

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
