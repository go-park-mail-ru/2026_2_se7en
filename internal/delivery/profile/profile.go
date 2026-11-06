package profile

import (
	"errors"
	"io"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
	usecaseprofile "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/profile"
)

type ProfileDeliveryImpl struct {
	uc ProfileUsecase
}

func NewProfileDelivery(uc ProfileUsecase) *ProfileDeliveryImpl {
	return &ProfileDeliveryImpl{uc: uc}
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
		Email:     req.Email,
		Nickname:  req.Nickname,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Bio:       req.Bio,
	}

	err := d.uc.UpdateProfile(r.Context(), userID, params)
	if err != nil {
		if errors.Is(err, apperrors.ErrEmailTaken) || errors.Is(err, apperrors.ErrNicknameTaken) {
			return apperrors.NewConflict(err.Error(), nil)
		}

		return apperrors.NewInternalError("Something went wrong", err)
	}

	user, prof, err := d.uc.GetProfile(r.Context(), userID)
	if err != nil {
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

	if err := r.ParseMultipartForm(tools.MaxAvatarSize); err != nil {
		apperrors.WriteError(w, apperrors.NewBadRequest("File too large", nil))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		apperrors.WriteError(w, apperrors.NewBadRequest("File is required", nil))
		return
	}

	defer file.Close()

	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		apperrors.WriteError(w, apperrors.NewInternalError("Failed to read file", err))
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		apperrors.WriteError(w, apperrors.NewInternalError("Failed to seek file", err))
		return
	}

	contentType := http.DetectContentType(buffer)
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid file type", nil))
		return
	}

	upload := &models.FileUpload{
		Reader:      file,
		Filename:    header.Filename,
		Size:        header.Size,
		ContentType: contentType,
	}

	url, err := d.uc.UploadAvatar(r.Context(), userID, upload)
	if err != nil {
		apperrors.WriteError(w, apperrors.NewInternalError("Failed to upload avatar", err))
		return
	}

	tools.WriteJSON(w, http.StatusOK, map[string]string{"icon_url": url})
}
