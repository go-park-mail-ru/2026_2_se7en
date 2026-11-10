package chat

import (
	"github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/httperrors"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/google/uuid"
)

// HandleCreateGroup godoc
// @Summary      Создать групповой чат
// @Description  Создаёт группу и добавляет указанных участников
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     SessionCookie
// @Param        request body CreateGroupRequest true "Данные группы"
// @Success      201 {object} ChatDetailResponse
// @Failure      400 {object} httperrors.ErrorResponse
// @Failure      401 {object} httperrors.ErrorResponse
// @Failure      404 {object} httperrors.ErrorResponse
// @Failure      500 {object} httperrors.ErrorResponse
// @Router       /chats/groups [post]
func (d *ChatDeliveryImpl) HandleCreateGroup(w http.ResponseWriter, r *http.Request, req *CreateGroupRequest) *httperrors.HTTPError {

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		return httperrors.NewUnauthorized("Authorization required")
	}

	for _, memberID := range req.Members {
		if memberID == userID {
			return httperrors.NewBadRequest("Validation failed", []httperrors.ErrorDetail{{Field: "members", Reason: "Creator must not be included in members"}})
		}
	}

	result, err := d.chatUsecase.CreateGroup(r.Context(), userID, req.Name, req.Description, req.Members)
	if err != nil {
		d.log.ErrorContext(r.Context(), "HandleCreateGroup failed", logger.Error(err))

		return mapChatError(err)
	}

	event := models.Event{
		Type:    models.EventChatCreated,
		Payload: map[string]any{"chat_id": result.Chat.ID},
	}

	d.broadcaster.BroadcastToUser(userID, event)
	for _, memberID := range req.Members {
		d.broadcaster.BroadcastToUser(memberID, event)
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
// @Failure      400 {object} httperrors.ErrorResponse
// @Failure      401 {object} httperrors.ErrorResponse
// @Failure      403 {object} httperrors.ErrorResponse
// @Failure      404 {object} httperrors.ErrorResponse
// @Failure      500 {object} httperrors.ErrorResponse
// @Router       /chats/{id} [patch]
func (d *ChatDeliveryImpl) HandleUpdateGroup(w http.ResponseWriter, r *http.Request, req *UpdateGroupRequest) *httperrors.HTTPError {

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		return httperrors.NewUnauthorized("Authorization required")
	}

	chatID, err := uuid.Parse(r.PathValue("id"))
	if err != nil || chatID == uuid.Nil {
		if err != nil {
			d.log.ErrorContext(r.Context(), "HandleUpdateGroup failed", logger.Error(err))
		}

		return httperrors.NewBadRequest("Invalid id", []httperrors.ErrorDetail{{Field: "id", Reason: "id must be a valid UUID"}})
	}

	result, usecaseErr := d.chatUsecase.UpdateGroup(r.Context(), userID, chatID, req.Name, req.Description, req.SetDescription)
	if usecaseErr != nil {
		return mapChatError(usecaseErr)
	}

	d.broadcaster.BroadcastToChat(chatID, models.Event{
		Type:    models.EventChatUpdated,
		Payload: map[string]any{"chat_id": chatID},
	})
	d.broadcastChatActivity(r.Context(), chatID)
	tools.WriteJSON(w, http.StatusOK, ChatDetailToResponse(result))
	return nil
}

// HandleUploadGroupAvatar godoc
// @Summary      Загрузить или заменить аватар группы
// @Description  Только создатель или администратор. Файл JPEG, PNG или WebP до 5 МиБ; icon_url — путь относительно адреса backend.
// @Tags         chats
// @Accept       multipart/form-data
// @Produce      json
// @Security     SessionCookie
// @Param        id path string true "ID группы (UUID)" format(uuid)
// @Param        file formData file true "Аватар"
// @Success      200 {object} ChatDetailResponse
// @Failure      400 {object} httperrors.ErrorResponse
// @Failure      401 {object} httperrors.ErrorResponse
// @Failure      403 {object} httperrors.ErrorResponse
// @Failure      404 {object} httperrors.ErrorResponse
// @Failure      500 {object} httperrors.ErrorResponse
// @Router       /chats/{id}/avatar [put]
func (d *ChatDeliveryImpl) HandleUploadGroupAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httperrors.WriteError(w, httperrors.NewUnauthorized("Authorization required"))
		return
	}
	chatID, err := uuid.Parse(r.PathValue("id"))
	if err != nil || chatID == uuid.Nil {
		httperrors.WriteError(w, httperrors.NewBadRequest("Invalid id", nil))
		return
	}
	upload, closeFile, err := tools.ExtractAvatar(w, r)
	if err != nil {
		httperrors.WriteError(w, httperrors.NewBadRequest("Invalid avatar file", nil))
		return
	}
	defer closeFile()
	detail, err := d.chatUsecase.UploadGroupAvatar(r.Context(), userID, chatID, upload)
	if err != nil {
		httperrors.WriteError(w, mapChatError(err))
		return
	}
	d.broadcaster.BroadcastToChat(chatID, models.Event{Type: models.EventChatUpdated, Payload: map[string]any{"chat_id": chatID}})
	d.broadcastChatActivity(r.Context(), chatID)
	tools.WriteJSON(w, http.StatusOK, ChatDetailToResponse(detail))
}

// HandleGetListMembers godoc
// @Summary      Участники чата
// @Description  Возвращает активных участников чата с ролями, профилями и статусом онлайн
// @Tags         chats
// @Produce      json
// @Security     SessionCookie
// @Param        id path string true "ID чата (UUID)" format(uuid)
// @Success      200 {object} ListMembersResponse
// @Failure      400 {object} httperrors.ErrorResponse
// @Failure      401 {object} httperrors.ErrorResponse
// @Failure      404 {object} httperrors.ErrorResponse
// @Failure      500 {object} httperrors.ErrorResponse
// @Router       /chats/{id}/members [get]
func (d *ChatDeliveryImpl) HandleGetListMembers(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httperrors.WriteError(w, httperrors.NewUnauthorized("Authorization required"))
		return
	}

	chatID, err := uuid.Parse(r.PathValue("id"))
	if err != nil || chatID == uuid.Nil {
		if err != nil {
			d.log.ErrorContext(r.Context(), "HandleGetListMembers failed", logger.Error(err))
		}

		httperrors.WriteError(w, httperrors.NewBadRequest("Invalid id", []httperrors.ErrorDetail{{Field: "id", Reason: "id must be a valid UUID"}}))
		return
	}

	result, err := d.chatUsecase.GetListMembers(r.Context(), userID, chatID)
	if err != nil {
		d.log.ErrorContext(r.Context(), "HandleGetListMembers failed", logger.Error(err))

		httperrors.WriteError(w, mapChatError(err))
		return
	}

	response := ChatMembersToResponse(result)
	for i, member := range result {
		response.Items[i].IsOnline = d.broadcaster.IsOnline(member.UserID)
	}
	tools.WriteJSON(w, http.StatusOK, response)
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
// @Failure      400 {object} httperrors.ErrorResponse
// @Failure      401 {object} httperrors.ErrorResponse
// @Failure      403 {object} httperrors.ErrorResponse
// @Failure      404 {object} httperrors.ErrorResponse
// @Failure      409 {object} httperrors.ErrorResponse
// @Failure      500 {object} httperrors.ErrorResponse
// @Router       /chats/{id}/members [post]
func (d *ChatDeliveryImpl) HandleAddMember(w http.ResponseWriter, r *http.Request, req *AddMemberRequest) *httperrors.HTTPError {

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		return httperrors.NewUnauthorized("Authorization required")
	}

	if req.UserID == userID {
		return httperrors.NewBadRequest("Validation failed", []httperrors.ErrorDetail{{Field: "user_id", Reason: "Cannot add yourself"}})
	}

	chatID, err := uuid.Parse(r.PathValue("id"))
	if err != nil || chatID == uuid.Nil {
		if err != nil {
			d.log.ErrorContext(r.Context(), "HandleAddMember failed", logger.Error(err))
		}

		return httperrors.NewBadRequest("Invalid id", []httperrors.ErrorDetail{{Field: "id", Reason: "id must be a valid UUID"}})
	}

	result, usecaseErr := d.chatUsecase.AddMember(r.Context(), userID, chatID, req.UserID)
	if usecaseErr != nil {
		return mapChatError(usecaseErr)
	}
	memberResponse := ChatMemberToResponse(*result)
	memberResponse.IsOnline = d.broadcaster.IsOnline(result.UserID)

	event := models.Event{
		Type: models.EventMemberAdded,
		Payload: map[string]any{
			"chat_id": chatID,
			"member":  memberResponse,
		},
	}

	d.broadcaster.BroadcastToChat(chatID, event)
	d.broadcaster.BroadcastToUser(req.UserID, event)
	d.broadcastChatActivity(r.Context(), chatID)

	tools.WriteJSON(w, http.StatusCreated, memberResponse)
	return nil
}

// HandleUpdateMemberRole godoc
// @Summary      Изменить роль участника группы
// @Description  Только создатель группы может назначить участника администратором или вернуть его к роли участника
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     SessionCookie
// @Param        id path string true "ID группы (UUID)" format(uuid)
// @Param        user_id path string true "ID участника (UUID)" format(uuid)
// @Param        request body UpdateMemberRoleRequest true "Новая роль участника"
// @Success      200 {object} ChatMemberResponse
// @Failure      400 {object} httperrors.ErrorResponse
// @Failure      401 {object} httperrors.ErrorResponse
// @Failure      403 {object} httperrors.ErrorResponse
// @Failure      404 {object} httperrors.ErrorResponse
// @Failure      500 {object} httperrors.ErrorResponse
// @Router       /chats/{id}/members/{user_id}/role [patch]
func (d *ChatDeliveryImpl) HandleUpdateMemberRole(w http.ResponseWriter, r *http.Request, req *UpdateMemberRoleRequest) *httperrors.HTTPError {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		return httperrors.NewUnauthorized("Authorization required")
	}
	if req.Role != "admin" && req.Role != "member" {
		return httperrors.NewBadRequest("Validation failed", []httperrors.ErrorDetail{{Field: "role", Reason: "Role must be admin or member"}})
	}

	chatID, err := uuid.Parse(r.PathValue("id"))
	if err != nil || chatID == uuid.Nil {
		return httperrors.NewBadRequest("Invalid id", []httperrors.ErrorDetail{{Field: "id", Reason: "id must be a valid UUID"}})
	}
	targetID, err := uuid.Parse(r.PathValue("user_id"))
	if err != nil || targetID == uuid.Nil {
		return httperrors.NewBadRequest("Invalid user_id", []httperrors.ErrorDetail{{Field: "user_id", Reason: "user_id must be a valid UUID"}})
	}

	result, usecaseErr := d.chatUsecase.UpdateMemberRole(r.Context(), userID, chatID, targetID, req.Role)
	if usecaseErr != nil {
		return mapChatError(usecaseErr)
	}

	memberResponse := ChatMemberToResponse(*result)
	memberResponse.IsOnline = d.broadcaster.IsOnline(result.UserID)
	d.broadcaster.BroadcastToChat(chatID, models.Event{
		Type: models.EventMemberUpdated,
		Payload: map[string]any{
			"chat_id": chatID,
			"user_id": targetID,
			"role":    result.Role,
		},
	})
	d.broadcastChatActivity(r.Context(), chatID)
	tools.WriteJSON(w, http.StatusOK, memberResponse)
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
// @Failure      400 {object} httperrors.ErrorResponse
// @Failure      401 {object} httperrors.ErrorResponse
// @Failure      403 {object} httperrors.ErrorResponse
// @Failure      404 {object} httperrors.ErrorResponse
// @Failure      500 {object} httperrors.ErrorResponse
// @Router       /chats/{id}/members/{user_id} [delete]
func (d *ChatDeliveryImpl) HandleRemoveMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httperrors.WriteError(w, httperrors.NewUnauthorized("Authorization required"))
		return
	}

	chatID, err := uuid.Parse(r.PathValue("id"))
	if err != nil || chatID == uuid.Nil {
		if err != nil {
			d.log.ErrorContext(r.Context(), "HandleRemoveMember failed", logger.Error(err))
		}

		httperrors.WriteError(w, httperrors.NewBadRequest("Invalid id", []httperrors.ErrorDetail{{Field: "id", Reason: "id must be a valid UUID"}}))
		return
	}

	targetID, err := uuid.Parse(r.PathValue("user_id"))
	if err != nil || targetID == uuid.Nil {
		if err != nil {
			d.log.ErrorContext(r.Context(), "HandleRemoveMember failed", logger.Error(err))
		}

		httperrors.WriteError(w, httperrors.NewBadRequest("Invalid user_id", []httperrors.ErrorDetail{{Field: "user_id", Reason: "user_id must be a valid UUID"}}))
		return
	}

	if err := d.chatUsecase.RemoveMember(r.Context(), userID, chatID, targetID); err != nil {
		d.log.ErrorContext(r.Context(), "HandleRemoveMember failed", logger.Error(err))

		httperrors.WriteError(w, mapChatError(err))
		return
	}

	d.broadcastMemberRemoved(chatID, targetID)
	d.broadcastChatActivity(r.Context(), chatID)
	d.broadcaster.BroadcastToUsers([]uuid.UUID{targetID}, models.Event{Type: models.EventChatActivity, Payload: map[string]any{"chat_id": chatID}})

	w.WriteHeader(http.StatusNoContent)
}

// HandleLeaveGroup godoc
// @Summary      Выйти из группы
// @Description  Удаляет текущего пользователя из группы
// @Tags         chats
// @Security     SessionCookie
// @Param        id path string true "ID группы (UUID)" format(uuid)
// @Success      204 "Выход из группы выполнен"
// @Failure      400 {object} httperrors.ErrorResponse
// @Failure      401 {object} httperrors.ErrorResponse
// @Failure      403 {object} httperrors.ErrorResponse
// @Failure      404 {object} httperrors.ErrorResponse
// @Failure      500 {object} httperrors.ErrorResponse
// @Router       /chats/{id}/leave [delete]
func (d *ChatDeliveryImpl) HandleLeaveGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httperrors.WriteError(w, httperrors.NewUnauthorized("Authorization required"))
		return
	}

	chatID, err := uuid.Parse(r.PathValue("id"))
	if err != nil || chatID == uuid.Nil {
		if err != nil {
			d.log.ErrorContext(r.Context(), "HandleLeaveGroup failed", logger.Error(err))
		}

		httperrors.WriteError(w, httperrors.NewBadRequest("Invalid id", []httperrors.ErrorDetail{{Field: "id", Reason: "id must be a valid UUID"}}))
		return
	}

	if err := d.chatUsecase.RemoveMember(r.Context(), userID, chatID, userID); err != nil {
		d.log.ErrorContext(r.Context(), "HandleLeaveGroup failed", logger.Error(err))

		httperrors.WriteError(w, mapChatError(err))
		return
	}

	d.broadcastMemberRemoved(chatID, userID)
	d.broadcastChatActivity(r.Context(), chatID)
	d.broadcaster.BroadcastToUsers([]uuid.UUID{userID}, models.Event{Type: models.EventChatActivity, Payload: map[string]any{"chat_id": chatID}})
	w.WriteHeader(http.StatusNoContent)
}

func (d *ChatDeliveryImpl) broadcastMemberRemoved(chatID, userID uuid.UUID) {
	d.broadcaster.RevokeUserFromChat(userID, chatID)
	event := models.Event{
		Type: models.EventMemberRemoved,
		Payload: map[string]any{
			"chat_id": chatID,
			"user_id": userID,
		},
	}

	d.broadcaster.BroadcastToChat(chatID, event)
	d.broadcaster.BroadcastToUser(userID, event)
}
