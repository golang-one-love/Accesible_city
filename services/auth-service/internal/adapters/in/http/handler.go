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
	g.POST("/self-promote", h.SelfPromote, h.authMiddleware.RequireAuth())
	g.POST("/apply", h.ApplyForRole, h.authMiddleware.RequireAuth())
	g.GET("/applications", h.ListApplications, h.authMiddleware.RequireAuth(), h.authMiddleware.RequireRoles(entity.RoleModerator, entity.RoleAdmin))
	g.POST("/applications/:id/approve", h.ApproveApplication, h.authMiddleware.RequireAuth(), h.authMiddleware.RequireRoles(entity.RoleModerator, entity.RoleAdmin))
	g.POST("/applications/:id/reject", h.RejectApplication, h.authMiddleware.RequireAuth(), h.authMiddleware.RequireRoles(entity.RoleModerator, entity.RoleAdmin))
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

func (h *AuthHandler) SelfPromote(c echo.Context) error {
	var req SelfPromoteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	user, err := h.authUseCase.SelfPromote(GetUserID(c), entity.Role(req.Role))
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(http.StatusOK, toUserResponse(user))
}

func (h *AuthHandler) ApplyForRole(c echo.Context) error {
	var req ApplyRoleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	app, err := h.authUseCase.ApplyForRole(GetUserID(c), entity.Role(req.Role), req.Comment)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(http.StatusCreated, toApplicationResponse(app))
}

func (h *AuthHandler) ListApplications(c echo.Context) error {
	apps, err := h.authUseCase.ListApplications(GetUserRole(c))
	if err != nil {
		return h.handleError(c, err)
	}

	response := ApplicationsResponse{Applications: make([]RoleApplicationResponse, 0, len(apps))}
	for _, app := range apps {
		response.Applications = append(response.Applications, toApplicationResponse(app))
	}
	return c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) ApproveApplication(c echo.Context) error {
	return h.reviewApplication(c, true)
}

func (h *AuthHandler) RejectApplication(c echo.Context) error {
	return h.reviewApplication(c, false)
}

func (h *AuthHandler) reviewApplication(c echo.Context, approve bool) error {
	app, err := h.authUseCase.ReviewApplication(GetUserID(c), GetUserRole(c), c.Param("id"), approve)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(http.StatusOK, toApplicationResponse(app))
}

func (h *AuthHandler) Validate(c echo.Context) error {	claims, err := h.authUseCase.ValidateAccessToken(c.Request().Header.Get("Authorization")[7:])
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
	case service.ErrApplicationNotFound:
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
	case service.ErrApplicationExists:
		return c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
	case service.ErrApplicationInvalidStatus:
		return c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
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

func toApplicationResponse(app *entity.RoleApplication) RoleApplicationResponse {
	reviewedAt := ""
	if !app.ReviewedAt.IsZero() {
		reviewedAt = app.ReviewedAt.Format(time.RFC3339)
	}
	return RoleApplicationResponse{
		ID:            app.ID,
		UserID:        app.UserID,
		RequestedRole: string(app.RequestedRole),
		Comment:       app.Comment,
		Status:        string(app.Status),
		CreatedAt:     app.CreatedAt.Format(time.RFC3339),
		ReviewedAt:    reviewedAt,
	}
}