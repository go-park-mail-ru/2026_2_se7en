package profile

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
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
// @Failure 400 {object} apperrors.ErrorResponse
// @Failure 401 {object} apperrors.ErrorResponse
// @Failure 404 {object} apperrors.ErrorResponse
// @Failure 500 {object} apperrors.ErrorResponse
// @Router       /profile [get]
func (d *ProfileDeliveryImpl) HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Authorization required"))
		return
	}

	user, prof, err := d.uc.GetProfile(r.Context(), userID)
	if err != nil {
		d.log.ErrorContext(r.Context(), "HandleGetProfile failed", logger.Error(err))

		if errors.Is(err, apperrors.ErrUserNotFound) {
			apperrors.WriteError(w, apperrors.NewNotFound("User not found"))
			return
		}
		apperrors.WriteError(w, apperrors.NewInternalError("Something went wrong", err))
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
// @Failure 400 {object} apperrors.ErrorResponse
// @Failure 401 {object} apperrors.ErrorResponse
// @Failure 409 {object} apperrors.ErrorResponse
// @Failure 500 {object} apperrors.ErrorResponse
// @Router       /profile [patch]
func (d *ProfileDeliveryImpl) HandleUpdateProfile(w http.ResponseWriter, r *http.Request, req *UpdateProfileRequest) *apperrors.AppError {

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		return apperrors.NewUnauthorized("Authorization required")
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

		if errors.Is(err, apperrors.ErrEmailTaken) || errors.Is(err, apperrors.ErrNicknameTaken) {
			return apperrors.NewConflict(err.Error(), nil)
		}

		return apperrors.NewInternalError("Something went wrong", err)
	}

	user, prof, err := d.uc.GetProfile(r.Context(), userID)
	if err != nil {
		d.log.ErrorContext(r.Context(), "HandleUpdateProfile failed", logger.Error(err))

		return apperrors.NewInternalError("Something went wrong", err)
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
// @Success      200  {object} map[string]string
// @Failure 400 {object} apperrors.ErrorResponse
// @Failure 401 {object} apperrors.ErrorResponse
// @Failure 500 {object} apperrors.ErrorResponse
// @Router       /profile/avatar [put]
func (d *ProfileDeliveryImpl) HandleUploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Authorization required"))
		return
	}

	upload, closeFunc, appErr := extractAndValidateAvatar(r)
	if appErr != nil {
		d.log.ErrorContext(r.Context(), "HandleUploadAvatar failed", logger.Error(appErr))

		apperrors.WriteError(w, appErr)
		return
	}
	defer closeFunc()

	url, err := d.uc.UploadAvatar(r.Context(), userID, upload)
	if err != nil {
		d.log.ErrorContext(r.Context(), "HandleUploadAvatar failed", logger.Error(err))

		apperrors.WriteError(w, apperrors.NewInternalError("Failed to upload avatar", err))
		return
	}

	tools.WriteJSON(w, http.StatusOK, map[string]string{"icon_url": url})
}

func extractAndValidateAvatar(r *http.Request) (*models.FileUpload, func(), *apperrors.AppError) {
	if err := r.ParseMultipartForm(tools.MaxAvatarSize); err != nil {
		slog.Default().ErrorContext(r.Context(), "extractAndValidateAvatar failed", logger.Error(err))

		return nil, nil, apperrors.NewBadRequest("File too large", nil)
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		slog.Default().ErrorContext(r.Context(), "extractAndValidateAvatar failed", logger.Error(err))

		return nil, nil, apperrors.NewBadRequest("File is required", nil)
	}

	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		slog.Default().ErrorContext(r.Context(), "extractAndValidateAvatar failed", logger.Error(err))

		_ = file.Close()
		return nil, nil, apperrors.NewInternalError("Failed to read file", err)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		slog.Default().ErrorContext(r.Context(), "extractAndValidateAvatar failed", logger.Error(err))

		_ = file.Close()
		return nil, nil, apperrors.NewInternalError("Failed to seek file", err)
	}

	contentType := http.DetectContentType(buffer)
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		_ = file.Close()
		return nil, nil, apperrors.NewBadRequest("Invalid file type", nil)
	}

	upload := &models.FileUpload{
		Reader:      file,
		Filename:    header.Filename,
		Size:        header.Size,
		ContentType: contentType,
	}

	return upload, func() { _ = file.Close() }, nil
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
// @Failure 400 {object} apperrors.ErrorResponse
// @Failure 401 {object} apperrors.ErrorResponse
// @Failure 500 {object} apperrors.ErrorResponse
// @Router       /profile/password [patch]
func (d *ProfileDeliveryImpl) HandleUpdatePassword(w http.ResponseWriter, r *http.Request, req *UpdatePasswordRequest) *apperrors.AppError {

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		return apperrors.NewUnauthorized("Authorization required")
	}

	if err := d.uc.UpdatePassword(r.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		d.log.ErrorContext(r.Context(), "HandleUpdatePassword failed", logger.Error(err))

		if errors.Is(err, apperrors.ErrInvalidCredentials) {
			return apperrors.NewBadRequest("Invalid old password", []apperrors.ErrorDetail{
				{Field: "old_password", Reason: "invalid_old_password"},
			})
		}
		return apperrors.NewInternalError("Something went wrong", err)
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
