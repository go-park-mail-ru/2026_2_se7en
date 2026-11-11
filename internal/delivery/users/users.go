package users

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/httperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/google/uuid"
)

type OnlineChecker interface {
	IsOnline(userID uuid.UUID) bool
}

type UsersDeliveryImpl struct {
	log           *slog.Logger
	usersUsecase  UsersUsecase
	onlineChecker OnlineChecker
}

func NewUsersDelivery(usersUsecase UsersUsecase, onlineChecker OnlineChecker, baseLog *slog.Logger) *UsersDeliveryImpl {
	return &UsersDeliveryImpl{
		log:           logger.Scope(baseLog, logger.LayerDelivery, logger.ModuleUser),
		usersUsecase:  usersUsecase,
		onlineChecker: onlineChecker,
	}
}

// HandleGetUsersList godoc
// @Summary      Поиск пользователей по никнейму
// @Description  Ищет активных пользователей по части никнейма без учёта регистра; текущий пользователь не включается
// @Tags         users
// @Produce      json
// @Security     SessionCookie
// @Param        q query string true "Часть никнейма"
// @Param        limit query int false "Размер страницы (1-50)" default(20)
// @Param        offset query int false "Смещение страницы" default(0)
// @Success      200 {object} UsersListResponse
// @Failure      400 {object} httperrors.ErrorResponse
// @Failure      401 {object} httperrors.ErrorResponse
// @Failure      500 {object} httperrors.ErrorResponse
// @Router       /users [get]
func (d *UsersDeliveryImpl) HandleGetUsersList(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httperrors.WriteError(w, httperrors.NewUnauthorized("Invalid session"))
		return
	}

	nickname := strings.TrimSpace(r.URL.Query().Get("q"))
	limit, offset := 20, 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 50 {
			httperrors.WriteError(w, httperrors.NewBadRequest("Invalid limit", []httperrors.ErrorDetail{{Field: "limit", Reason: "Limit must be between 1 and 50"}}))
			return
		}
		limit = parsed
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			httperrors.WriteError(w, httperrors.NewBadRequest("Invalid offset", []httperrors.ErrorDetail{{Field: "offset", Reason: "Offset must be >= 0"}}))
			return
		}
		offset = parsed
	}

	response := UsersListResponse{Items: []UserSearchResponse{}}
	if nickname == "" {
		tools.WriteJSON(w, http.StatusOK, response)
		return
	}

	profiles, err := d.usersUsecase.GetUsersList(r.Context(), userID, limit+1, offset, nickname)
	if err != nil {
		d.log.ErrorContext(r.Context(), "HandleGetUsersList failed", logger.Error(err))
		httperrors.WriteError(w, httperrors.NewInternalError("Something went wrong", err))
		return
	}
	if len(profiles) > limit {
		profiles = profiles[:limit]
		nextOffset := offset + limit
		response.NextOffset = &nextOffset
	}
	response.Items = UserProfilesToResponse(profiles, func(id uuid.UUID) bool {
		return d.onlineChecker != nil && d.onlineChecker.IsOnline(id)
	})
	tools.WriteJSON(w, http.StatusOK, response)
}
