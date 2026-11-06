package contact

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
)

type ContactDeliveryImpl struct {
	contactUsecase ContactUsecase
}

func NewContactDelivery(contactUsecase ContactUsecase) *ContactDeliveryImpl {
	return &ContactDeliveryImpl{
		contactUsecase: contactUsecase,
	}
}

// HandleGetContactsList godoc
// @Summary      Список контактов
// @Description  Возвращает список контактов текущего пользователя с профилями и статусом онлайн
// @Tags         contacts
// @Produce      json
// @Security     SessionCookie
// @Success      200  {array} ContactResponse
// @Failure      401  {object} apperrors.ErrorResponse
// @Failure      500  {object} apperrors.ErrorResponse
// @Router       /contacts [get]
func (d *ContactDeliveryImpl) HandleGetContactsList(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Invalid session"))
		return
	}

	contactsList, err := d.contactUsecase.GetContactsList(r.Context(), userID)
	if err != nil {
		apperrors.WriteError(w, apperrors.NewInternalError("Something went wrong", err))
		return
	}

	response := ContactsToResponse(contactsList)
	tools.WriteJSON(w, http.StatusOK, response)
}

// HandleCreateContact godoc
// @Summary      Добавить контакт
// @Description  Добавляет пользователя в список контактов по его никнейму
// @Tags         contacts
// @Accept       json
// @Produce      json
// @Security     SessionCookie
// @Param        request body AddContactRequest true "Данные для добавления контакта"
// @Success      201  "Контакт успешно добавлен"
// @Failure      400  {object} apperrors.ErrorResponse
// @Failure      401  {object} apperrors.ErrorResponse
// @Failure      404  {object} apperrors.ErrorResponse
// @Failure      409  {object} apperrors.ErrorResponse
// @Failure      500  {object} apperrors.ErrorResponse
// @Router       /contacts [post]
func (d *ContactDeliveryImpl) HandleCreateContact(w http.ResponseWriter, r *http.Request, req *CreateContactRequest) *apperrors.AppError {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		return apperrors.NewUnauthorized("Invalid session")
	}

	err := d.contactUsecase.CreateContact(r.Context(), userID, req.Nickname, req.Pseudonym)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrUserNotFound):
			return apperrors.NewNotFound("User not found")
		case errors.Is(err, apperrors.ErrContactAlreadyExists):
			return apperrors.NewConflict("User already in contacts", nil)
		case errors.Is(err, apperrors.ErrCannotAddSelfAsContact):
			return apperrors.NewBadRequest("Cannot add self as contact", nil)
		default:
			return apperrors.NewInternalError("Something went wrong", err)
		}
	}

	w.WriteHeader(http.StatusCreated)
	return nil
}

// HandleDeleteContact godoc
// @Summary      Удалить контакт
// @Description  Удаляет пользователя из списка контактов
// @Tags         contacts
// @Accept       json
// @Produce      json
// @Security     SessionCookie
// @Param        request body RemoveContactRequest true "Данные для удаления контакта"
// @Success      204  "Контакт успешно удален"
// @Failure      400  {object} apperrors.ErrorResponse
// @Failure      401  {object} apperrors.ErrorResponse
// @Failure      404  {object} apperrors.ErrorResponse
// @Failure      500  {object} apperrors.ErrorResponse
// @Router       /contacts [delete]
func (d *ContactDeliveryImpl) HandleDeleteContact(w http.ResponseWriter, r *http.Request, req *DeleteContactRequest) *apperrors.AppError {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		return apperrors.NewUnauthorized("Invalid session")
	}

	err := d.contactUsecase.DeleteContact(r.Context(), userID, req.ContactUserID)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrContactNotFound):
			return apperrors.NewNotFound("Контакт не найден")
		default:
			return apperrors.NewInternalError("Something went wrong", err)
		}
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
