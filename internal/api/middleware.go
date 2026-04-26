package api

import (
	"net/http"

	"github.com/bubbletrack/server/internal/domain"
	"github.com/labstack/echo/v4"
)

func MockUserMiddleware(provider domain.UserProvider) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID := provider.GetCurrentUserID(c.Request().Context())
			c.Set("user_id", userID)
			return next(c)
		}
	}
}

func ErrorHandler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		msg := "internal server error"

		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			if m, ok := he.Message.(string); ok {
				msg = m
			}
		}

		if _, ok := err.(*domain.NotFoundError); ok {
			code = http.StatusNotFound
			msg = err.Error()
		}

		if _, ok := err.(*domain.ValidationError); ok {
			code = http.StatusBadRequest
			msg = err.Error()
		}

		c.JSON(code, echo.Map{"error": msg})
	}
}
