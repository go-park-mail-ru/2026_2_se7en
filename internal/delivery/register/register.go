package register

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
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
// @Failure      400  {object} apperrors.ErrorResponse
// @Failure      409  {object} apperrors.ErrorResponse
// @Failure      500  {object} apperrors.ErrorResponse
// @Router       /auth/register [post]
func (d *RegisterDeliveryImpl) HandleRegister(w http.ResponseWriter, r *http.Request, req *RegisterRequest) *apperrors.AppError {

	userModel := RegisterRequestToModel(req)

	registerRes, err := d.registerUsecase.Register(r.Context(), userModel, req.Password)
	if err != nil {
		d.log.ErrorContext(r.Context(), "HandleRegister failed", logger.Error(err))

		if errors.Is(err, apperrors.ErrEmailTaken) {
			return apperrors.NewConflictWithCode(apperrors.ErrEmailTaken.Code, "Email already taken", []apperrors.ErrorDetail{
				{Field: "email", Reason: "Email already exists"},
			})
		}
		if errors.Is(err, apperrors.ErrNicknameTaken) {
			return apperrors.NewConflictWithCode(apperrors.ErrNicknameTaken.Code, "Nickname already taken", []apperrors.ErrorDetail{
				{Field: "nickname", Reason: "Nickname already exists"},
			})
		}
		return apperrors.NewInternalError("Something went wrong", err)
	}

	response := ModelToRegisterResponse(registerRes.User)

	middleware.SetSessionCookie(w, registerRes.SessionID.String())

	tools.WriteJSON(w, http.StatusCreated, response)
	return nil
}
