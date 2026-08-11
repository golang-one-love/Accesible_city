package out

import (
	"github.com/accessible-path/auth-service/internal/domain/entity"
	"github.com/accessible-path/auth-service/internal/domain/valueobject"
)

type UserRepository interface {
	Create(user *entity.User) error
	GetByEmail(email valueobject.Email) (*entity.User, error)
	GetByID(id string) (*entity.User, error)
	Update(user *entity.User) error
}

type TokenGenerator interface {
	GenerateAccessToken(user *entity.User) (string, error)
	GenerateRefreshToken(user *entity.User) (string, error)
	ValidateAccessToken(token string) (*TokenClaims, error)
	ValidateRefreshToken(token string) (*TokenClaims, error)
}

type TokenClaims struct {
	UserID string
	Email  string
	Role   entity.Role
	Type   string
	Exp    int64
}