package chat

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
	"github.com/google/uuid"
)

type ChatDeliveryImpl struct {
	chatUsecase ChatUsecase
}

func NewChatDelivery(chatUsecase ChatUsecase) *ChatDeliveryImpl {
	return &ChatDeliveryImpl{chatUsecase: chatUsecase}
}

// HandleListUserChats godoc
// @Summary      Список чатов пользователя
// @Description  Возвращает список чатов с поддержкой пагинации и фильтрации по типу
// @Tags         chats
// @Produce      json
// @Security     SessionCookie
// @Param        limit  query    int     false  "Размер списка (1-50)"  default(20)
// @Param        offset query    int     false  "Смещение от начала"     default(0)
// @Param        type   query    string  false  "Фильтр по типу чата"    Enums(dialog, group, channel)
// @Success      200  {object} ListChatsResponse
// @Failure 400 {object} apperrors.ErrorResponse
// @Failure 401 {object} apperrors.ErrorResponse
// @Failure 409 {object} apperrors.ErrorResponse
// @Failure 500 {object} apperrors.ErrorResponse
// @Router       /chats [get]
func (d *ChatDeliveryImpl) HandleListUserChats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Authorization required"))
		return
	}

	q := r.URL.Query()

	limit, err := parseQueryInt(q, "limit")
	if err != nil {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid limit format", []apperrors.ErrorDetail{{Field: "limit", Reason: err.Error()}}))
		return
	}
	if _, present := q["limit"]; !present {
		limit = 20
	}
	if limit < 1 || limit > 50 {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid limit", []apperrors.ErrorDetail{{Field: "limit", Reason: "Limit must be between 1 and 50"}}))
		return
	}

	offset, err := parseQueryInt(q, "offset")
	if err != nil {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid offset format", []apperrors.ErrorDetail{{Field: "offset", Reason: err.Error()}}))
		return
	}
	if offset < 0 {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid offset", []apperrors.ErrorDetail{{Field: "offset", Reason: "Offset must be >= 0"}}))
		return
	}

	chatType := ""
	if vals, present := q["type"]; present {
		if len(vals) != 1 || vals[0] == "" {
			apperrors.WriteError(w, apperrors.NewBadRequest("Invalid type format", []apperrors.ErrorDetail{{Field: "type", Reason: "Type must be a single non-empty value"}}))
			return
		}
		chatType = vals[0]
	}
	if chatType != "" && chatType != "dialog" && chatType != "group" && chatType != "channel" {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid type", []apperrors.ErrorDetail{{Field: "type", Reason: "Invalid chat type"}}))
		return
	}

	items, err := d.chatUsecase.ListUserChats(r.Context(), userID, limit, offset, chatType)
	if err != nil {
		apperrors.WriteError(w, mapChatError(err))
		return
	}

	response := ChatItemsToResponse(items)
	tools.WriteJSON(w, http.StatusOK, response)
}

// HandleGetChatDetail godoc
// @Summary      Детали чата для пользователя
// @Description  Возвращает чат и роль текущего пользователя в нём
// @Tags         chats
// @Produce      json
// @Security     SessionCookie
// @Param        id path string true "ID чата (UUID)" format(uuid)
// @Success      200  {object} ChatDetailResponse
// @Failure 400 {object} apperrors.ErrorResponse
// @Failure 401 {object} apperrors.ErrorResponse
// @Failure 404 {object} apperrors.ErrorResponse
// @Failure 500 {object} apperrors.ErrorResponse
// @Router       /chats/{id} [get]
func (d *ChatDeliveryImpl) HandleGetChatDetail(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Authorization required"))
		return
	}
	chatID, err := uuid.Parse(r.PathValue("id"))
	if err != nil || chatID == uuid.Nil {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid id", []apperrors.ErrorDetail{{Field: "id", Reason: "id must be a valid UUID"}}))
		return
	}
	result, err := d.chatUsecase.GetChatDetail(r.Context(), userID, chatID)
	if err != nil {
		apperrors.WriteError(w, mapChatError(err))
		return
	}
	tools.WriteJSON(w, http.StatusOK, ChatDetailToResponse(result))
}

// HandleGetOrCreateDialog godoc
// @Summary      Получить или создать диалог
// @Description  Возвращает существующий диалог с пользователем или создаёт новый
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     SessionCookie
// @Param        request body GetOrCreateDialogRequest true "Данные для создания"
// @Success      200 {object} ChatDetailResponse
// @Failure      400 {object} apperrors.ErrorResponse
// @Failure      401 {object} apperrors.ErrorResponse
// @Failure      404 {object} apperrors.ErrorResponse
// @Failure      500 {object} apperrors.ErrorResponse
// @Router       /chats/dialogs [post]
func (d *ChatDeliveryImpl) HandleGetOrCreateDialog(w http.ResponseWriter, r *http.Request, req *GetOrCreateDialogRequest) *apperrors.AppError {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		return apperrors.NewUnauthorized("Authorization required")
	}
	if userID == req.PeerID {
		return apperrors.NewBadRequest("Validation failed", []apperrors.ErrorDetail{{Field: "peer_id", Reason: "Cannot create a dialog with yourself"}})
	}
	result, err := d.chatUsecase.GetOrCreateDialog(r.Context(), userID, req.PeerID)
	if err != nil {
		return mapChatError(err)
	}
	tools.WriteJSON(w, http.StatusOK, ChatDetailToResponse(result))
	return nil
}

// HandleHideDialog godoc
// @Summary      Скрыть диалог
// @Description  Убирает текущего пользователя из диалога
// @Tags         chats
// @Security     SessionCookie
// @Param        id path string true "ID диалога (UUID)" format(uuid)
// @Success      204 "Диалог скрыт"
// @Failure      400 {object} apperrors.ErrorResponse
// @Failure      401 {object} apperrors.ErrorResponse
// @Failure      404 {object} apperrors.ErrorResponse
// @Failure      500 {object} apperrors.ErrorResponse
// @Router       /chats/{id}/hide [post]
func (d *ChatDeliveryImpl) HandleHideDialog(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Authorization required"))
		return
	}
	chatID, err := uuid.Parse(r.PathValue("id"))
	if err != nil || chatID == uuid.Nil {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid id", []apperrors.ErrorDetail{{Field: "id", Reason: "id must be a valid UUID"}}))
		return
	}
	if err := d.chatUsecase.HideDialog(r.Context(), userID, chatID); err != nil {
		apperrors.WriteError(w, mapChatError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HandleCreateGroup godoc
// @Summary      Создать групповой чат
// @Description  Создаёт группу и добавляет указанных участников
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     SessionCookie
// @Param        request body CreateGroupRequest true "Данные группы"
// @Success      201 {object} ChatDetailResponse
// @Failure      400 {object} apperrors.ErrorResponse
// @Failure      401 {object} apperrors.ErrorResponse
// @Failure      404 {object} apperrors.ErrorResponse
// @Failure      500 {object} apperrors.ErrorResponse
// @Router       /chats/groups [post]
func (d *ChatDeliveryImpl) HandleCreateGroup(w http.ResponseWriter, r *http.Request, req *CreateGroupRequest) *apperrors.AppError {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		return apperrors.NewUnauthorized("Authorization required")
	}
	for _, memberID := range req.Members {
		if memberID == userID {
			return apperrors.NewBadRequest("Validation failed", []apperrors.ErrorDetail{{Field: "members", Reason: "Creator must not be included in members"}})
		}
	}
	result, err := d.chatUsecase.CreateGroup(r.Context(), userID, req.Name, req.Description, req.Members)
	if err != nil {
		return mapChatError(err)
	}
	tools.WriteJSON(w, http.StatusCreated, ChatDetailToResponse(result))
	return nil
}

// HandleUpdateGroup godoc
// @Summary      Изменить групповой чат
// @Description  Обновляет название и/или описание группы; set_description разрешает изменить или очистить описание
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     SessionCookie
// @Param        id path string true "ID группы (UUID)" format(uuid)
// @Param        request body UpdateGroupRequest true "Изменяемые поля"
// @Success      200 {object} ChatDetailResponse
// @Failure      400 {object} apperrors.ErrorResponse
// @Failure      401 {object} apperrors.ErrorResponse
// @Failure      403 {object} apperrors.ErrorResponse
// @Failure      404 {object} apperrors.ErrorResponse
// @Failure      500 {object} apperrors.ErrorResponse
// @Router       /chats/{id} [patch]
func (d *ChatDeliveryImpl) HandleUpdateGroup(w http.ResponseWriter, r *http.Request, req *UpdateGroupRequest) *apperrors.AppError {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		return apperrors.NewUnauthorized("Authorization required")
	}
	chatID, err := uuid.Parse(r.PathValue("id"))
	if err != nil || chatID == uuid.Nil {
		return apperrors.NewBadRequest("Invalid id", []apperrors.ErrorDetail{{Field: "id", Reason: "id must be a valid UUID"}})
	}
	result, usecaseErr := d.chatUsecase.UpdateGroup(r.Context(), userID, chatID, req.Name, req.Description, req.SetDescription)
	if usecaseErr != nil {
		return mapChatError(usecaseErr)
	}
	tools.WriteJSON(w, http.StatusOK, ChatDetailToResponse(result))
	return nil
}

// HandleDeleteGroup godoc
// @Summary      Удалить групповой чат
// @Description  Удаляет группу; действие доступно её создателю
// @Tags         chats
// @Security     SessionCookie
// @Param        id path string true "ID группы (UUID)" format(uuid)
// @Success      204 "Группа удалена"
// @Failure      400 {object} apperrors.ErrorResponse
// @Failure      401 {object} apperrors.ErrorResponse
// @Failure      403 {object} apperrors.ErrorResponse
// @Failure      404 {object} apperrors.ErrorResponse
// @Failure      500 {object} apperrors.ErrorResponse
// @Router       /chats/{id} [delete]
func (d *ChatDeliveryImpl) HandleDeleteGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Authorization required"))
		return
	}
	chatID, err := uuid.Parse(r.PathValue("id"))
	if err != nil || chatID == uuid.Nil {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid id", []apperrors.ErrorDetail{{Field: "id", Reason: "id must be a valid UUID"}}))
		return
	}
	if err := d.chatUsecase.DeleteGroup(r.Context(), userID, chatID); err != nil {
		apperrors.WriteError(w, mapChatError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HandleListMembers godoc
// @Summary      Участники чата
// @Description  Возвращает активных участников чата с ролями и профилями
// @Tags         chats
// @Produce      json
// @Security     SessionCookie
// @Param        id path string true "ID чата (UUID)" format(uuid)
// @Success      200 {object} ListMembersResponse
// @Failure      400 {object} apperrors.ErrorResponse
// @Failure      401 {object} apperrors.ErrorResponse
// @Failure      404 {object} apperrors.ErrorResponse
// @Failure      500 {object} apperrors.ErrorResponse
// @Router       /chats/{id}/members [get]
func (d *ChatDeliveryImpl) HandleListMembers(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Authorization required"))
		return
	}
	chatID, err := uuid.Parse(r.PathValue("id"))
	if err != nil || chatID == uuid.Nil {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid id", []apperrors.ErrorDetail{{Field: "id", Reason: "id must be a valid UUID"}}))
		return
	}
	result, err := d.chatUsecase.ListMembers(r.Context(), userID, chatID)
	if err != nil {
		apperrors.WriteError(w, mapChatError(err))
		return
	}
	tools.WriteJSON(w, http.StatusOK, ChatMembersToResponse(result))
}

// HandleAddMember godoc
// @Summary      Добавить участника в группу
// @Description  Добавляет пользователя в группу и возвращает его данные
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     SessionCookie
// @Param        id path string true "ID группы (UUID)" format(uuid)
// @Param        request body AddMemberRequest true "ID нового участника"
// @Success      201 {object} ChatMemberResponse
// @Failure      400 {object} apperrors.ErrorResponse
// @Failure      401 {object} apperrors.ErrorResponse
// @Failure      403 {object} apperrors.ErrorResponse
// @Failure      404 {object} apperrors.ErrorResponse
// @Failure      409 {object} apperrors.ErrorResponse
// @Failure      500 {object} apperrors.ErrorResponse
// @Router       /chats/{id}/members [post]
func (d *ChatDeliveryImpl) HandleAddMember(w http.ResponseWriter, r *http.Request, req *AddMemberRequest) *apperrors.AppError {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		return apperrors.NewUnauthorized("Authorization required")
	}
	if req.UserID == userID {
		return apperrors.NewBadRequest("Validation failed", []apperrors.ErrorDetail{{Field: "user_id", Reason: "Cannot add yourself"}})
	}
	chatID, err := uuid.Parse(r.PathValue("id"))
	if err != nil || chatID == uuid.Nil {
		return apperrors.NewBadRequest("Invalid id", []apperrors.ErrorDetail{{Field: "id", Reason: "id must be a valid UUID"}})
	}
	result, usecaseErr := d.chatUsecase.AddMember(r.Context(), userID, chatID, req.UserID)
	if usecaseErr != nil {
		return mapChatError(usecaseErr)
	}
	tools.WriteJSON(w, http.StatusCreated, ChatMemberToResponse(*result))
	return nil
}

// HandleRemoveMember godoc
// @Summary      Удалить участника из группы
// @Description  Удаляет участника группы или позволяет пользователю выйти из неё
// @Tags         chats
// @Security     SessionCookie
// @Param        id path string true "ID группы (UUID)" format(uuid)
// @Param        user_id path string true "ID участника (UUID)" format(uuid)
// @Success      204 "Участник удалён"
// @Failure      400 {object} apperrors.ErrorResponse
// @Failure      401 {object} apperrors.ErrorResponse
// @Failure      403 {object} apperrors.ErrorResponse
// @Failure      404 {object} apperrors.ErrorResponse
// @Failure      500 {object} apperrors.ErrorResponse
// @Router       /chats/{id}/members/{user_id} [delete]
func (d *ChatDeliveryImpl) HandleRemoveMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Authorization required"))
		return
	}
	chatID, err := uuid.Parse(r.PathValue("id"))
	if err != nil || chatID == uuid.Nil {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid id", []apperrors.ErrorDetail{{Field: "id", Reason: "id must be a valid UUID"}}))
		return
	}
	targetID, err := uuid.Parse(r.PathValue("user_id"))
	if err != nil || targetID == uuid.Nil {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid user_id", []apperrors.ErrorDetail{{Field: "user_id", Reason: "user_id must be a valid UUID"}}))
		return
	}
	if err := d.chatUsecase.RemoveMember(r.Context(), userID, chatID, targetID); err != nil {
		apperrors.WriteError(w, mapChatError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HandleListMessages godoc
// @Summary      Сообщения чата
// @Description  Возвращает сообщения чата с пагинацией по ID предыдущего сообщения
// @Tags         chats
// @Produce      json
// @Security     SessionCookie
// @Param        id path string true "ID чата (UUID)" format(uuid)
// @Param        limit query int false "Число сообщений (1-50)" default(20)
// @Param        before query string false "ID сообщения для пагинации (UUID)" format(uuid)
// @Success      200 {object} ListMessagesResponse
// @Failure      400 {object} apperrors.ErrorResponse
// @Failure      401 {object} apperrors.ErrorResponse
// @Failure      404 {object} apperrors.ErrorResponse
// @Failure      500 {object} apperrors.ErrorResponse
// @Router       /chats/{id}/messages [get]
func (d *ChatDeliveryImpl) HandleListMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Authorization required"))
		return
	}
	chatID, err := uuid.Parse(r.PathValue("id"))
	if err != nil || chatID == uuid.Nil {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid id", []apperrors.ErrorDetail{{Field: "id", Reason: "id must be a valid UUID"}}))
		return
	}
	q := r.URL.Query()
	limit := 20
	if _, present := q["limit"]; present {
		parsed, err := parseQueryInt(q, "limit")
		if err != nil {
			apperrors.WriteError(w, apperrors.NewBadRequest("Invalid limit format", nil))
			return
		}
		limit = parsed
	}
	if limit < 1 || limit > 50 {
		apperrors.WriteError(w, apperrors.NewBadRequest("Validation failed", []apperrors.ErrorDetail{{Field: "limit", Reason: "Limit must be between 1 and 50"}}))
		return
	}
	beforeID, err := parseQueryUUID(q, "before")
	if err != nil {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid before format", nil))
		return
	}
	var before *uuid.UUID
	if _, present := q["before"]; present {
		before = &beforeID
	}
	result, err := d.chatUsecase.ListMessages(r.Context(), userID, chatID, limit, before)
	if err != nil {
		apperrors.WriteError(w, mapChatError(err))
		return
	}
	tools.WriteJSON(w, http.StatusOK, MessagesToResponse(result))
}

// HandleCreateMessage godoc
// @Summary      Создать сообщение
// @Description  Отправляет текст или стикер в чат и возвращает созданное сообщение
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     SessionCookie
// @Param        id path string true "ID чата (UUID)" format(uuid)
// @Param        request body CreateMessageRequest true "Данные сообщения"
// @Success      201 {object} MessageResponse
// @Failure      400 {object} apperrors.ErrorResponse
// @Failure      401 {object} apperrors.ErrorResponse
// @Failure      404 {object} apperrors.ErrorResponse
// @Failure      500 {object} apperrors.ErrorResponse
// @Router       /chats/{id}/messages [post]
func (d *ChatDeliveryImpl) HandleCreateMessage(w http.ResponseWriter, r *http.Request, req *CreateMessageRequest) *apperrors.AppError {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		return apperrors.NewUnauthorized("Authorization required")
	}
	chatID, err := uuid.Parse(r.PathValue("id"))
	if err != nil || chatID == uuid.Nil {
		return apperrors.NewBadRequest("Invalid id", []apperrors.ErrorDetail{{Field: "id", Reason: "id must be a valid UUID"}})
	}
	result, usecaseErr := d.chatUsecase.CreateMessage(r.Context(), userID, chatID, req.ReplyToMessageID, req.StickerID, req.Content, req.Type)
	if usecaseErr != nil {
		return mapChatError(usecaseErr)
	}
	tools.WriteJSON(w, http.StatusCreated, MessageToResponse(*result))
	return nil
}

func parseQueryInt(query url.Values, key string) (int, error) {
	vals, present := query[key]
	if !present {
		return 0, nil
	}
	if len(vals) != 1 {
		return 0, apperrors.ErrInvalidChatLimit
	}

	val, err := strconv.Atoi(vals[0])
	if err != nil {
		return 0, err
	}
	return val, nil
}

func parseQueryUUID(query url.Values, key string) (uuid.UUID, error) {
	vals, present := query[key]
	if !present {
		return uuid.Nil, nil
	}
	if len(vals) != 1 {
		return uuid.Nil, apperrors.ErrInvalidUUID
	}

	val, err := uuid.Parse(vals[0])
	if err != nil || val == uuid.Nil {
		return uuid.Nil, apperrors.ErrInvalidUUID
	}
	return val, nil
}

func mapChatError(err error) *apperrors.AppError {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	switch {
	case errors.Is(err, apperrors.ErrInvalidChatLimit):
		return apperrors.NewBadRequest("Invalid limit", []apperrors.ErrorDetail{{Field: "limit", Reason: "Limit must be >= 1"}})
	case errors.Is(err, apperrors.ErrInvalidChatOffset):
		return apperrors.NewBadRequest("Invalid offset", []apperrors.ErrorDetail{{Field: "offset", Reason: "Offset must be >= 0"}})
	case errors.Is(err, apperrors.ErrInvalidChatType):
		return apperrors.NewBadRequest("Invalid type", []apperrors.ErrorDetail{{Field: "type", Reason: "Invalid chat type"}})
	case errors.Is(err, apperrors.ErrChatNotFound), errors.Is(err, apperrors.ErrChatMemberNotFound), errors.Is(err, apperrors.ErrUserNotFound):
		return apperrors.NewNotFound("Resource not found")
	case errors.Is(err, apperrors.ErrChatPermissionDenied):
		return apperrors.NewForbidden("No permissions")
	case errors.Is(err, apperrors.ErrInvalidChatOperation), errors.Is(err, apperrors.ErrMessageNotFound):
		return apperrors.NewBadRequest("Validation failed", nil)
	case errors.Is(err, apperrors.ErrChatMemberExists):
		return apperrors.NewConflict("Resource already exists", nil)
	default:
		return apperrors.NewInternalError("Something went wrong", err)
	}
}
