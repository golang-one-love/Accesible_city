package usecase

import (
	"github.com/accessible-path/auth-service/internal/domain/entity"
	"github.com/accessible-path/auth-service/internal/domain/service"
	"github.com/accessible-path/auth-service/internal/ports/in"
)

type authUseCase struct {
	authService *service.AuthService
}

func NewAuthUseCase(authService *service.AuthService) in.AuthUseCase {
	return &authUseCase{authService: authService}
}

func (u *authUseCase) Register(email, password string, role entity.Role) (*entity.User, string, string, error) {
	return u.authService.Register(email, password, role)
}

func (u *authUseCase) Login(email, password string) (*entity.User, string, string, error) {
	return u.authService.Login(email, password)
}

func (u *authUseCase) Refresh(refreshToken string) (string, string, error) {
	return u.authService.Refresh(refreshToken)
}

func (u *authUseCase) ValidateAccessToken(token string) (*service.TokenClaims, error) {
	return u.authService.ValidateAccessToken(token)
}

func (u *authUseCase) GetUserByID(id string) (*entity.User, error) {
	return u.authService.GetUserByID(id)
}

func (u *authUseCase) UpdateUserRole(userID string, role entity.Role) error {
	return u.authService.UpdateUserRole(userID, role)
}