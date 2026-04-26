package api

import (
	"net/http"
	"strings"

	"github.com/bubbletrack/server/internal/application"
	"github.com/bubbletrack/server/internal/domain"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authUC *application.AuthUseCase
}

func NewAuthHandler(authUC *application.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

func (h *AuthHandler) RegisterRoutes(e *echo.Echo) {
	auth := e.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.RefreshToken)
	auth.POST("/logout", h.Logout)
	auth.GET("/me", h.Me)
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req domain.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	resp, err := h.authUC.Register(c.Request().Context(), &req)
	if err != nil {
		if _, ok := err.(*domain.ValidationError); ok {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "registration failed"})
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req domain.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	resp, err := h.authUC.Login(c.Request().Context(), &req)
	if err != nil {
		if _, ok := err.(*domain.ValidationError); ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "login failed"})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req domain.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	resp, err := h.authUC.RefreshToken(c.Request().Context(), &req)
	if err != nil {
		if _, ok := err.(*domain.ValidationError); ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "refresh failed"})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c echo.Context) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.Bind(&req)
	_ = h.authUC.Logout(c.Request().Context(), req.RefreshToken)
	return c.JSON(http.StatusOK, map[string]string{"status": "logged out"})
}

func (h *AuthHandler) Me(c echo.Context) error {
	userID := c.Get("user_id").(string)
	account, err := h.authUC.GetCurrentUser(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
	}
	return c.JSON(http.StatusOK, account)
}

var publicAuthPaths = map[string]bool{
	"/auth/register": true,
	"/auth/login":    true,
	"/auth/refresh":  true,
}

func JWTMiddleware(authUC *application.AuthUseCase) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if publicAuthPaths[c.Path()] {
				return next(c)
			}

			if c.Path() == "/ws" || c.Path() == "/api/health" {
				return next(c)
			}

			header := c.Request().Header.Get("Authorization")
			if header == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid authorization format"})
			}

			userID, err := authUC.ValidateAccessToken(parts[1])
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
			}

			c.Set("user_id", userID)
			return next(c)
		}
	}
}
