package http

import (
	"net/http"
	"time"

	"github.com/accessible-path/auth-service/internal/domain/entity"
	"github.com/accessible-path/auth-service/internal/domain/service"
	"github.com/accessible-path/auth-service/internal/ports/in"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authUseCase   in.AuthUseCase
	authMiddleware *AuthMiddleware
}

func NewAuthHandler(authUseCase in.AuthUseCase, authMiddleware *AuthMiddleware) *AuthHandler {
	return &AuthHandler{
		authUseCase:    authUseCase,
		authMiddleware: authMiddleware,
	}
}

func (h *AuthHandler) RegisterRoutes(g *echo.Group) {
	g.POST("/register", h.Register)
	g.POST("/login", h.Login)
	g.POST("/refresh", h.Refresh)
	g.GET("/validate", h.Validate, h.authMiddleware.RequireAuth())
	g.GET("/profile", h.Profile, h.authMiddleware.RequireAuth())
	g.PATCH("/profile", h.UpdateProfile, h.authMiddleware.RequireAuth())
	g.POST("/change-password", h.ChangePassword, h.authMiddleware.RequireAuth())
	g.GET("/users", h.ListUsers, h.authMiddleware.RequireAuth(), h.authMiddleware.RequireRoles(entity.RoleModerator, entity.RoleAdmin))
	g.PATCH("/users/:id/role", h.UpdateUserRole, h.authMiddleware.RequireAuth(), h.authMiddleware.RequireRoles(entity.RoleModerator, entity.RoleAdmin))
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	user, accessToken, refreshToken, err := h.authUseCase.Register(req.Email, req.Password, entity.RoleUser)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusCreated, AuthResponse{
		User:         toUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	user, accessToken, refreshToken, err := h.authUseCase.Login(req.Email, req.Password)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, AuthResponse{
		User:         toUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	var req RefreshRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	accessToken, refreshToken, err := h.authUseCase.Refresh(req.RefreshToken)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *AuthHandler) Profile(c echo.Context) error {
	user, err := h.authUseCase.GetUserByID(GetUserID(c))
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(http.StatusOK, toUserResponse(user))
}

func (h *AuthHandler) UpdateProfile(c echo.Context) error {
	var req UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}
	if req.Nickname == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: service.ErrInvalidNickname.Error()})
	}

	user, err := h.authUseCase.UpdateProfile(GetUserID(c), req.Nickname)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(http.StatusOK, toUserResponse(user))
}

func (h *AuthHandler) ChangePassword(c echo.Context) error {
	var req ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	if err := h.authUseCase.ChangePassword(GetUserID(c), req.OldPassword, req.NewPassword); err != nil {
		return h.handleError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *AuthHandler) ListUsers(c echo.Context) error {
	users, err := h.authUseCase.ListUsers()
	if err != nil {
		return h.handleError(c, err)
	}

	response := UsersResponse{Users: make([]UserResponse, 0, len(users))}
	for _, user := range users {
		response.Users = append(response.Users, toUserResponse(user))
	}
	return c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) UpdateUserRole(c echo.Context) error {
	var req UpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	targetID := c.Param("id")
	if err := h.authUseCase.UpdateUserRole(GetUserID(c), GetUserRole(c), targetID, entity.Role(req.Role)); err != nil {
		return h.handleError(c, err)
	}

	user, err := h.authUseCase.GetUserByID(targetID)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(http.StatusOK, toUserResponse(user))
}

func (h *AuthHandler) Validate(c echo.Context) error {
	claims, err := h.authUseCase.ValidateAccessToken(c.Request().Header.Get("Authorization")[7:])
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
	}

	return c.JSON(http.StatusOK, ValidateResponse{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   string(claims.Role),
		Exp:    claims.Exp,
	})
}

func (h *AuthHandler) handleError(c echo.Context, err error) error {
	switch err {
	case service.ErrUserAlreadyExists:
		return c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
	case service.ErrInvalidCredentials, service.ErrUserInactive:
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
	case service.ErrInvalidRole, service.ErrInvalidNickname:
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	case service.ErrWrongPassword:
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	case service.ErrForbidden:
		return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
	case service.ErrTokenExpired, service.ErrTokenInvalid:
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
	case service.ErrUserNotFound:
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
	default:
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
}

func toUserResponse(user *entity.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email.String(),
		Nickname:  user.Nickname,
		Role:      string(user.Role),
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}
}