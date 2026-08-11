package valueobject

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidEmail = errors.New("invalid email format")
	emailRegex      = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

type Email string

func NewEmail(value string) (Email, error) {
	v := strings.TrimSpace(strings.ToLower(value))
	if v == "" {
		return "", ErrInvalidEmail
	}
	if !emailRegex.MatchString(v) {
		return "", ErrInvalidEmail
	}
	if len(v) > 254 {
		return "", ErrInvalidEmail
	}
	return Email(v), nil
}

func (e Email) String() string {
	return string(e)
}

func (e Email) IsValid() bool {
	return emailRegex.MatchString(string(e))
}