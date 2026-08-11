package valueobject

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong  = errors.New("password must not exceed 72 characters")
	ErrInvalidHash      = errors.New("invalid password hash")
)

type Password string

func NewPassword(value string) (Password, error) {
	if len(value) < 8 {
		return "", ErrPasswordTooShort
	}
	if len(value) > 72 {
		return "", ErrPasswordTooLong
	}
	return Password(value), nil
}

func (p Password) Hash(cost int) (PasswordHash, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(p), cost)
	if err != nil {
		return "", err
	}
	return PasswordHash(bytes), nil
}

func (p Password) Compare(hash PasswordHash) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(p))
}

type PasswordHash string

func NewPasswordHash(hash string) (PasswordHash, error) {
	if len(hash) == 0 {
		return "", ErrInvalidHash
	}
	return PasswordHash(hash), nil
}

func (h PasswordHash) String() string {
	return string(h)
}

func (h PasswordHash) Bytes() []byte {
	return []byte(h)
}