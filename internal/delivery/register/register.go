package register

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/httperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/apperrors"
)

type RegisterDeliveryImpl struct {
	log *slog.Logger

	registerUsecase RegisterUsecase
}

func NewRegisterDelivery(registerUsecase RegisterUsecase, baseLog *slog.Logger) *RegisterDeliveryImpl {
	return &RegisterDeliveryImpl{
		log: logger.Scope(baseLog, logger.LayerDelivery, logger.ModuleRegister),

		registerUsecase: registerUsecase,
	}
}

// HandleRegister godoc
// @Summary      Регистрация пользователя
// @Description  Регистрация нового пользователя с автоматическим входом
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "Данные для регистрации"
// @Success      201  {object} RegisterResponse
// @Failure      400  {object} httperrors.ErrorResponse
// @Failure      409  {object} httperrors.ErrorResponse
// @Failure      500  {object} httperrors.ErrorResponse
// @Router       /auth/register [post]
func (d *RegisterDeliveryImpl) HandleRegister(w http.ResponseWriter, r *http.Request, req *RegisterRequest) *httperrors.HTTPError {

	userModel := RegisterRequestToModel(req)

	registerRes, err := d.registerUsecase.Register(r.Context(), userModel, req.Password)
	if err != nil {
		d.log.ErrorContext(r.Context(), "HandleRegister failed", logger.Error(err))

		if errors.Is(err, apperrors.ErrEmailTaken) {
			return httperrors.NewConflictWithCode("EMAIL_TAKEN", "Email already taken", []httperrors.ErrorDetail{
				{Field: "email", Reason: "Email already exists"},
			})
		}
		if errors.Is(err, apperrors.ErrNicknameTaken) {
			return httperrors.NewConflictWithCode("NICKNAME_TAKEN", "Nickname already taken", []httperrors.ErrorDetail{
				{Field: "nickname", Reason: "Nickname already exists"},
			})
		}
		if errors.Is(err, apperrors.ErrProfileRequired) {
			return httperrors.NewBadRequest("Profile is required", nil)
		}
		return httperrors.NewInternalError("Something went wrong", err)
	}

	response := ModelToRegisterResponse(registerRes.User)

	middleware.SetSessionCookie(w, r, registerRes.SessionID.String())

	tools.WriteJSON(w, http.StatusCreated, response)
	return nil
}
