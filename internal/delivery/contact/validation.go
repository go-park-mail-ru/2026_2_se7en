package contact

import (
	"github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/httperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/validation"
)

func (r *CreateContactRequest) IsValid() []httperrors.ErrorDetail {
	var details []httperrors.ErrorDetail

	if err := validation.ValidateNickname(r.Nickname); err != nil {
		details = append(details, *err)
	}

	if err := validation.ValidatePseudonym(r.Pseudonym); err != nil {
		details = append(details, *err)
	}

	return details
}
