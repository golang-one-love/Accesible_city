package in

import (
	"github.com/accessible-path/auth-service/internal/domain/entity"
	"github.com/accessible-path/auth-service/internal/domain/service"
)

type AuthUseCase interface {
	Register(email, password string, role entity.Role) (*entity.User, string, string, error)
	Login(email, password string) (*entity.User, string, string, error)
	Refresh(refreshToken string) (string, string, error)
	ValidateAccessToken(token string) (*service.TokenClaims, error)
	GetUserByID(id string) (*entity.User, error)
	UpdateUserRole(userID string, role entity.Role) error
}