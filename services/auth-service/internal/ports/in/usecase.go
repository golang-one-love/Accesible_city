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
	UpdateProfile(userID, nickname string) (*entity.User, error)
	ChangePassword(userID, oldPassword, newPassword string) error
	UpdateUserRole(actorID string, actorRole entity.Role, userID string, role entity.Role) error
	ListUsers() ([]*entity.User, error)
	SelfPromote(userID string, role entity.Role) (*entity.User, error)
	ApplyForRole(userID string, requestedRole entity.Role, comment string) (*entity.RoleApplication, error)
	ListApplications(actorRole entity.Role) ([]*entity.RoleApplication, error)
	ReviewApplication(actorID string, actorRole entity.Role, applicationID string, approve bool) (*entity.RoleApplication, error)
}