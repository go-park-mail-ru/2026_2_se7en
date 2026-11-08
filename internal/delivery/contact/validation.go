package contact

import (
	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
)

func (r *CreateContactRequest) IsValid() []apperrors.ErrorDetail {
	var details []apperrors.ErrorDetail

	if err := tools.ValidateNickname(r.Nickname); err != nil {
		details = append(details, *err)
	}

	if err := tools.ValidatePseudonym(r.Pseudonym); err != nil {
		details = append(details, *err)
	}

	return details
}
