package profile

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
	"github.com/google/uuid"
)

// HandleGetPublicProfile godoc
// @Summary Публичный профиль пользователя
// @Description Возвращает публичные данные профиля по ID пользователя без почты и телефона
// @Tags profile
// @Produce json
// @Security SessionCookie
// @Param id path string true "ID пользователя (UUID)" format(uuid)
// @Success 200 {object} PublicProfileResponse
// @Failure 400 {object} apperrors.ErrorResponse
// @Failure 401 {object} apperrors.ErrorResponse
// @Failure 404 {object} apperrors.ErrorResponse
// @Failure 500 {object} apperrors.ErrorResponse
// @Router /users/{id} [get]
func (d *ProfileDeliveryImpl) HandleGetPublicProfile(w http.ResponseWriter, r *http.Request) {
	if _, ok := middleware.UserIDFromContext(r.Context()); !ok {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Authorization required"))
		return
	}

	userID, err := uuid.Parse(r.PathValue("id"))
	if err != nil || userID == uuid.Nil {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid id", []apperrors.ErrorDetail{
			{Field: "id", Reason: "id must be a valid UUID"},
		}))
		return
	}

	user, prof, err := d.uc.GetProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			apperrors.WriteError(w, apperrors.NewNotFound("User not found"))
			return
		}
		apperrors.WriteError(w, apperrors.NewInternalError("Something went wrong", err))
		return
	}
	if user == nil || prof == nil {
		apperrors.WriteError(w, apperrors.NewNotFound("Profile not found"))
		return
	}

	tools.WriteJSON(w, http.StatusOK, ToPublicProfileResponse(user, prof))
}
