package chat

import (
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
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
	d.broadcaster.BroadcastToChat(chatID, models.Event{
		Type:    models.EventChatUpdated,
		Payload: map[string]any{"chat_id": chatID},
	})
	tools.WriteJSON(w, http.StatusOK, ChatDetailToResponse(result))
	return nil
}

// HandleGetListMembers godoc
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
func (d *ChatDeliveryImpl) HandleGetListMembers(w http.ResponseWriter, r *http.Request) {
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
	result, err := d.chatUsecase.GetListMembers(r.Context(), userID, chatID)
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

	event := models.Event{
		Type: models.EventMemberAdded,
		Payload: map[string]any{
			"chat_id": chatID,
			"member":  ChatMemberToResponse(*result),
		},
	}
	d.broadcaster.BroadcastToChat(chatID, event)
	d.broadcaster.BroadcastToUser(req.UserID, event)

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

	d.broadcastMemberRemoved(chatID, targetID)

	w.WriteHeader(http.StatusNoContent)
}

// HandleLeaveGroup godoc
// @Summary      Выйти из группы
// @Description  Удаляет текущего пользователя из группы
// @Tags         chats
// @Security     SessionCookie
// @Param        id path string true "ID группы (UUID)" format(uuid)
// @Success      204 "Выход из группы выполнен"
// @Failure      400 {object} apperrors.ErrorResponse
// @Failure      401 {object} apperrors.ErrorResponse
// @Failure      403 {object} apperrors.ErrorResponse
// @Failure      404 {object} apperrors.ErrorResponse
// @Failure      500 {object} apperrors.ErrorResponse
// @Router       /chats/{id}/leave [delete]
func (d *ChatDeliveryImpl) HandleLeaveGroup(w http.ResponseWriter, r *http.Request) {
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
	if err := d.chatUsecase.RemoveMember(r.Context(), userID, chatID, userID); err != nil {
		apperrors.WriteError(w, mapChatError(err))
		return
	}

	d.broadcastMemberRemoved(chatID, userID)
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

type CreateGroupRequest struct {
	Name        string      `json:"name"`
	Description *string     `json:"description,omitempty"`
	Members     []uuid.UUID `json:"members"`
}

type UpdateGroupRequest struct {
	Name           *string `json:"name,omitempty"`
	Description    *string `json:"description,omitempty"`
	SetDescription bool    `json:"set_description"`
}

type AddMemberRequest struct {
	UserID uuid.UUID `json:"user_id" format:"uuid"`
}

type MemberProfileResponse struct {
	ID        uuid.UUID  `json:"id" format:"uuid"`
	IconID    *uuid.UUID `json:"icon_id,omitempty" format:"uuid"`
	UserID    uuid.UUID  `json:"user_id" format:"uuid"`
	Nickname  string     `json:"nickname"`
	FirstName string     `json:"first_name"`
	LastName  *string    `json:"last_name,omitempty"`
	Bio       *string    `json:"bio,omitempty"`
	CreatedAt time.Time  `json:"created_at" format:"date-time"`
	UpdatedAt time.Time  `json:"updated_at" format:"date-time"`
}

type ChatMemberResponse struct {
	UserID  uuid.UUID              `json:"user_id" format:"uuid"`
	Role    string                 `json:"role"`
	Profile *MemberProfileResponse `json:"profile"`
}

type ListMembersResponse struct {
	Items []ChatMemberResponse `json:"items"`
}

func (r *CreateGroupRequest) IsValid() []apperrors.ErrorDetail {
	var details []apperrors.ErrorDetail
	if detail := tools.ValidateChatName(r.Name); detail != nil {
		details = append(details, *detail)
	}
	if detail := tools.ValidateChatDescription(r.Description); detail != nil {
		details = append(details, *detail)
	}
	seen := make(map[uuid.UUID]struct{}, len(r.Members))
	for _, id := range r.Members {
		if id == uuid.Nil {
			return append(details, apperrors.ErrorDetail{Field: "members", Reason: "invalid user_id"})
		}
		if _, exists := seen[id]; exists {
			return append(details, apperrors.ErrorDetail{Field: "members", Reason: "Members must be unique"})
		}
		seen[id] = struct{}{}
	}
	return details
}

func (r *UpdateGroupRequest) IsValid() []apperrors.ErrorDetail {
	if r.Name == nil && !r.SetDescription {
		return []apperrors.ErrorDetail{{Field: "name", Reason: "At least one field must be updated"}}
	}
	if r.Description != nil && !r.SetDescription {
		return []apperrors.ErrorDetail{{Field: "description", Reason: "Description update flag is required"}}
	}
	var details []apperrors.ErrorDetail
	if r.Name != nil {
		if detail := tools.ValidateChatName(*r.Name); detail != nil {
			details = append(details, *detail)
		}
	}
	if r.SetDescription {
		if detail := tools.ValidateChatDescription(r.Description); detail != nil {
			details = append(details, *detail)
		}
	}
	return details
}

func (r *AddMemberRequest) IsValid() []apperrors.ErrorDetail {
	if r.UserID == uuid.Nil {
		return []apperrors.ErrorDetail{{Field: "user_id", Reason: "invalid user_id"}}
	}
	return nil
}

func ChatMemberToResponse(member models.ChatMember) ChatMemberResponse {
	response := ChatMemberResponse{UserID: member.UserID, Role: member.Role}
	if member.Profile != nil {
		p := member.Profile
		response.Profile = &MemberProfileResponse{
			ID:        p.ID,
			IconID:    p.IconID,
			UserID:    p.UserID,
			Nickname:  p.Nickname,
			FirstName: p.FirstName,
			LastName:  p.LastName,
			Bio:       p.Bio,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		}
	}
	return response
}

func ChatMembersToResponse(members []models.ChatMember) ListMembersResponse {
	items := make([]ChatMemberResponse, 0, len(members))
	for _, member := range members {
		items = append(items, ChatMemberToResponse(member))
	}
	return ListMembersResponse{Items: items}
}
