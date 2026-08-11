package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/accessible-path/auth-service/internal/domain/valueobject"
)

type Role string

const (
	RoleUser         Role = "user"
	RoleVolunteer    Role = "volunteer"
	RoleModerator    Role = "moderator"
	RoleBusinessOwner Role = "business_owner"
	RoleAdmin        Role = "admin"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleUser, RoleVolunteer, RoleModerator, RoleBusinessOwner, RoleAdmin:
		return true
	default:
		return false
	}
}

func (r Role) CanModerate() bool {
	return r == RoleModerator || r == RoleAdmin
}

func (r Role) CanManagePOI() bool {
	return r == RoleBusinessOwner || r == RoleAdmin
}

type User struct {
	ID           string
	Email        valueobject.Email
	PasswordHash string
	Role         Role
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(email valueobject.Email, passwordHash string, role Role) *User {
	now := time.Now().UTC()
	return &User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (u *User) UpdateRole(role Role) {
	u.Role = role
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) Activate() {
	u.IsActive = true
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) Deactivate() {
	u.IsActive = false
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) UpdatePasswordHash(hash string) {
	u.PasswordHash = hash
	u.UpdatedAt = time.Now().UTC()
}