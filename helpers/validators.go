package helpers

import "regexp"

var (
	emailRegex       = regexp.MustCompile(`^[a-zA-Z0-9_]+@[a-zA-Z]\.[a-zA-Z]{2,}$`)
	phoneNumberRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	nicknameRegex    = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

func IsEmailValid(email string) bool {
	return emailRegex.MatchString(email)
}

func IsPhoneNumberValid(phone string) bool {
	return phoneNumberRegex.MatchString(phone)
}

func IsNicknameValid(nickname string) bool {
	return nicknameRegex.MatchString(nickname)
}
