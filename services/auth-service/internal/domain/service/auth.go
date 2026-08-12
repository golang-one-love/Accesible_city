package service

import (
	"errors"
	"strings"
	"time"

	"github.com/accessible-path/auth-service/internal/domain/entity"
	"github.com/accessible-path/auth-service/internal/domain/valueobject"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserInactive       = errors.New("user account is deactivated")
	ErrInvalidRole        = errors.New("invalid role")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenInvalid       = errors.New("token is invalid")
	ErrForbidden          = errors.New("insufficient permissions")
	ErrInvalidNickname    = errors.New("nickname must be 1-100 characters")
	ErrWrongPassword      = errors.New("current password is incorrect")
)

type AuthService struct {
	userRepo    UserRepository
	tokenGen    TokenGenerator
	bcryptCost  int
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

type UserRepository interface {
	Create(user *entity.User) error
	GetByEmail(email valueobject.Email) (*entity.User, error)
	GetByID(id string) (*entity.User, error)
	Update(user *entity.User) error
	List() ([]*entity.User, error)
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

func NewAuthService(
	userRepo UserRepository,
	tokenGen TokenGenerator,
	bcryptCost int,
	accessTTL, refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		tokenGen:   tokenGen,
		bcryptCost: bcryptCost,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (s *AuthService) Register(emailStr, passwordStr string, role entity.Role) (*entity.User, string, string, error) {
	email, err := valueobject.NewEmail(emailStr)
	if err != nil {
		return nil, "", "", err
	}

	if !role.IsValid() {
		return nil, "", "", ErrInvalidRole
	}

	existing, _ := s.userRepo.GetByEmail(email)
	if existing != nil {
		return nil, "", "", ErrUserAlreadyExists
	}

	password, err := valueobject.NewPassword(passwordStr)
	if err != nil {
		return nil, "", "", err
	}

	hash, err := password.Hash(s.bcryptCost)
	if err != nil {
		return nil, "", "", err
	}

	user := entity.NewUser(email, hash.String(), role)
	if err := s.userRepo.Create(user); err != nil {
		return nil, "", "", err
	}

	accessToken, err := s.tokenGen.GenerateAccessToken(user)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := s.tokenGen.GenerateRefreshToken(user)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *AuthService) Login(emailStr, passwordStr string) (*entity.User, string, string, error) {
	email, err := valueobject.NewEmail(emailStr)
	if err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, "", "", ErrUserInactive
	}

	hash, err := valueobject.NewPasswordHash(user.PasswordHash)
	if err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	password, err := valueobject.NewPassword(passwordStr)
	if err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	if err := password.Compare(hash); err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	accessToken, err := s.tokenGen.GenerateAccessToken(user)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := s.tokenGen.GenerateRefreshToken(user)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *AuthService) Refresh(refreshToken string) (string, string, error) {
	claims, err := s.tokenGen.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", ErrTokenInvalid
	}

	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return "", "", ErrUserNotFound
	}

	if !user.IsActive {
		return "", "", ErrUserInactive
	}

	newAccess, err := s.tokenGen.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	newRefresh, err := s.tokenGen.GenerateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	return newAccess, newRefresh, nil
}

func (s *AuthService) ValidateAccessToken(token string) (*TokenClaims, error) {
	return s.tokenGen.ValidateAccessToken(token)
}

func (s *AuthService) GetUserByID(id string) (*entity.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *AuthService) UpdateProfile(userID, nickname string) (*entity.User, error) {
	nickname = strings.TrimSpace(nickname)
	if len(nickname) == 0 || len(nickname) > 100 {
		return nil, ErrInvalidNickname
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	user.UpdateNickname(nickname)
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) ChangePassword(userID, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	hash, err := valueobject.NewPasswordHash(user.PasswordHash)
	if err != nil {
		return ErrWrongPassword
	}

	oldPass, err := valueobject.NewPassword(oldPassword)
	if err != nil {
		return ErrWrongPassword
	}

	if err := oldPass.Compare(hash); err != nil {
		return ErrWrongPassword
	}

	newPass, err := valueobject.NewPassword(newPassword)
	if err != nil {
		return err
	}

	newHash, err := newPass.Hash(s.bcryptCost)
	if err != nil {
		return err
	}

	user.UpdatePasswordHash(newHash.String())
	return s.userRepo.Update(user)
}

func (s *AuthService) UpdateUserRole(actorID string, actorRole entity.Role, userID string, role entity.Role) error {
	if !role.IsValid() {
		return ErrInvalidRole
	}

	if !actorRole.CanModerate() {
		return ErrForbidden
	}

	if actorRole == entity.RoleModerator && role == entity.RoleAdmin {
		return ErrForbidden
	}

	target, err := s.userRepo.GetByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	if actorRole == entity.RoleModerator && target.Role == entity.RoleAdmin {
		return ErrForbidden
	}

	target.UpdateRole(role)
	return s.userRepo.Update(target)
}

func (s *AuthService) ListUsers() ([]*entity.User, error) {
	return s.userRepo.List()
}