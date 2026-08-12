package http

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type UpdateProfileRequest struct {
	Nickname string `json:"nickname" validate:"required,max=100"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=72"`
}

type UpdateRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=user volunteer moderator business_owner admin"`
}

type SelfPromoteRequest struct {
	Role string `json:"role" validate:"required,oneof=volunteer"`
}

type ApplyRoleRequest struct {
	Role    string `json:"role" validate:"required,oneof=volunteer moderator"`
	Comment string `json:"comment"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Nickname  string `json:"nickname"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

type UsersResponse struct {
	Users []UserResponse `json:"users"`
}

type RoleApplicationResponse struct {
	ID            string `json:"id"`
	UserID        string `json:"user_id"`
	RequestedRole string `json:"requested_role"`
	Comment       string `json:"comment"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	ReviewedAt    string `json:"reviewed_at"`
}

type ApplicationsResponse struct {
	Applications []RoleApplicationResponse `json:"applications"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type ValidateResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Exp    int64  `json:"exp"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}