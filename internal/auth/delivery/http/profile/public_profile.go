package profile

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/infrastructure/http/httperrors"
	httpjson "github.com/go-park-mail-ru/2026_2_se7en/internal/infrastructure/http/json"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/infrastructure/http/middleware"
	"github.com/google/uuid"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/auth/model"
)

// HandleGetPublicProfile godoc
// @Summary Публичный профиль пользователя
// @Description Возвращает публичные данные профиля по ID пользователя без почты и телефона
// @Tags profile
// @Produce json
// @Security SessionCookie
// @Param id path string true "ID пользователя (UUID)" format(uuid)
// @Success 200 {object} PublicProfileResponse
// @Failure 400 {object} httperrors.ErrorResponse
// @Failure 401 {object} httperrors.ErrorResponse
// @Failure 404 {object} httperrors.ErrorResponse
// @Failure 500 {object} httperrors.ErrorResponse
// @Router /users/{id} [get]
func (d *ProfileDeliveryImpl) HandleGetPublicProfile(w http.ResponseWriter, r *http.Request) {
	if _, ok := middleware.UserIDFromContext(r.Context()); !ok {
		httperrors.WriteError(w, httperrors.NewUnauthorized("Authorization required"))
		return
	}

	userID, err := uuid.Parse(r.PathValue("id"))
	if err != nil || userID == uuid.Nil {
		httperrors.WriteError(w, httperrors.NewBadRequest("Invalid id", []httperrors.ErrorDetail{
			{Field: "id", Reason: "id must be a valid UUID"},
		}))
		return
	}

	user, prof, err := d.uc.GetProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			httperrors.WriteError(w, httperrors.NewNotFound("User not found"))
			return
		}
		httperrors.WriteError(w, httperrors.NewInternalError("Something went wrong", err))
		return
	}
	if user == nil || prof == nil {
		httperrors.WriteError(w, httperrors.NewNotFound("Profile not found"))
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, ToPublicProfileResponse(user, prof))
}
