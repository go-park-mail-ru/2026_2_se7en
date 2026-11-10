package profile

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/httperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/apperrors"
	usecaseprofile "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/profile"
)

type ProfileDeliveryImpl struct {
	log *slog.Logger

	uc ProfileUsecase
}

func NewProfileDelivery(uc ProfileUsecase, baseLog *slog.Logger) *ProfileDeliveryImpl {
	return &ProfileDeliveryImpl{
		log: logger.Scope(baseLog, logger.LayerDelivery, logger.ModuleProfile),
		uc:  uc}
}

// HandleGetProfile godoc
// @Summary      Профиль пользователя
// @Description  Получение информации о текущем пользователе и его профиле
// @Tags         profile
// @Produce      json
// @Security     SessionCookie
// @Success      200  {object} ProfileResponse
// @Failure 400 {object} httperrors.ErrorResponse
// @Failure 401 {object} httperrors.ErrorResponse
// @Failure 404 {object} httperrors.ErrorResponse
// @Failure 500 {object} httperrors.ErrorResponse
// @Router       /profile [get]
func (d *ProfileDeliveryImpl) HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httperrors.WriteError(w, httperrors.NewUnauthorized("Authorization required"))
		return
	}

	user, prof, err := d.uc.GetProfile(r.Context(), userID)
	if err != nil {
		d.log.ErrorContext(r.Context(), "HandleGetProfile failed", logger.Error(err))

		if errors.Is(err, apperrors.ErrUserNotFound) {
			httperrors.WriteError(w, httperrors.NewNotFound("User not found"))
			return
		}
		httperrors.WriteError(w, httperrors.NewInternalError("Something went wrong", err))
		return
	}

	tools.WriteJSON(w, http.StatusOK, ToProfileResponse(user, prof))
}

// HandleUpdateProfile godoc
// @Summary      Обновление профиля
// @Description  Обновление данных профиля и email пользователя
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     SessionCookie
// @Param        request body UpdateProfileRequest true "Данные для обновления"
// @Success      200  {object} ProfileResponse
// @Failure 400 {object} httperrors.ErrorResponse
// @Failure 401 {object} httperrors.ErrorResponse
// @Failure 409 {object} httperrors.ErrorResponse
// @Failure 500 {object} httperrors.ErrorResponse
// @Router       /profile [patch]
func (d *ProfileDeliveryImpl) HandleUpdateProfile(w http.ResponseWriter, r *http.Request, req *UpdateProfileRequest) *httperrors.HTTPError {

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		return httperrors.NewUnauthorized("Authorization required")
	}

	params := usecaseprofile.UpdateParams{
		Email:       req.Email,
		Nickname:    req.Nickname,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Bio:         req.Bio,
		PhoneNumber: req.PhoneNumber,
	}

	err := d.uc.UpdateProfile(r.Context(), userID, params)
	if err != nil {
		d.log.ErrorContext(r.Context(), "HandleUpdateProfile failed", logger.Error(err))

		if errors.Is(err, apperrors.ErrUserNotFound) || errors.Is(err, apperrors.ErrProfileNotFound) {
			return httperrors.NewNotFound("Profile not found")
		}
		if errors.Is(err, apperrors.ErrEmailTaken) || errors.Is(err, apperrors.ErrNicknameTaken) {
			return httperrors.NewConflict(err.Error(), nil)
		}

		return httperrors.NewInternalError("Something went wrong", err)
	}

	user, prof, err := d.uc.GetProfile(r.Context(), userID)
	if err != nil {
		d.log.ErrorContext(r.Context(), "HandleUpdateProfile failed", logger.Error(err))

		if errors.Is(err, apperrors.ErrUserNotFound) || errors.Is(err, apperrors.ErrProfileNotFound) {
			return httperrors.NewNotFound("Profile not found")
		}
		return httperrors.NewInternalError("Something went wrong", err)
	}

	tools.WriteJSON(w, http.StatusOK, ToProfileResponse(user, prof))

	return nil
}

// HandleUploadAvatar godoc
// @Summary      Загрузка аватарки
// @Description  Загрузка файла аватарки (multipart/form-data)
// @Tags         profile
// @Accept       multipart/form-data
// @Produce      json
// @Security     SessionCookie
// @Param        file formData file true "Файл аватарки"
// @Success      200  {object} AvatarResponse
// @Failure 400 {object} httperrors.ErrorResponse
// @Failure 401 {object} httperrors.ErrorResponse
// @Failure 500 {object} httperrors.ErrorResponse
// @Router       /profile/avatar [put]
func (d *ProfileDeliveryImpl) HandleUploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httperrors.WriteError(w, httperrors.NewUnauthorized("Authorization required"))
		return
	}

	upload, closeFunc, appErr := extractAndValidateAvatar(w, r)
	if appErr != nil {
		d.log.ErrorContext(r.Context(), "HandleUploadAvatar failed", logger.Error(appErr))

		httperrors.WriteError(w, appErr)
		return
	}
	defer closeFunc()

	file, err := d.uc.UploadAvatar(r.Context(), userID, upload)
	if err != nil {
		d.log.ErrorContext(r.Context(), "HandleUploadAvatar failed", logger.Error(err))

		if errors.Is(err, apperrors.ErrProfileNotFound) {
			httperrors.WriteError(w, httperrors.NewNotFound("Profile not found"))
			return
		}
		httperrors.WriteError(w, httperrors.NewInternalError("Failed to upload avatar", err))
		return
	}

	tools.WriteJSON(w, http.StatusOK, AvatarResponse{IconURL: file.URL})
}

func extractAndValidateAvatar(w http.ResponseWriter, r *http.Request) (*models.FileUpload, func(), *httperrors.HTTPError) {
	upload, closeFile, err := tools.ExtractAvatar(w, r)
	if err != nil {
		slog.Default().ErrorContext(r.Context(), "extractAndValidateAvatar failed", logger.Error(err))
		return nil, nil, httperrors.NewBadRequest("Invalid avatar file", nil)
	}
	return upload, closeFile, nil
}

// HandleUpdatePassword godoc
// @Summary      Смена пароля
// @Description  Меняет пароль пользователя после проверки старого
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     SessionCookie
// @Param        request body UpdatePasswordRequest true "Старый и новый пароль"
// @Success      204  "Пароль успешно изменён"
// @Failure 400 {object} httperrors.ErrorResponse
// @Failure 401 {object} httperrors.ErrorResponse
// @Failure 500 {object} httperrors.ErrorResponse
// @Router       /profile/password [patch]
func (d *ProfileDeliveryImpl) HandleUpdatePassword(w http.ResponseWriter, r *http.Request, req *UpdatePasswordRequest) *httperrors.HTTPError {

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		return httperrors.NewUnauthorized("Authorization required")
	}

	if err := d.uc.UpdatePassword(r.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		d.log.ErrorContext(r.Context(), "HandleUpdatePassword failed", logger.Error(err))

		if errors.Is(err, apperrors.ErrInvalidCredentials) {
			return httperrors.NewBadRequest("Invalid old password", []httperrors.ErrorDetail{
				{Field: "old_password", Reason: "invalid_old_password"},
			})
		}
		return httperrors.NewInternalError("Something went wrong", err)
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
