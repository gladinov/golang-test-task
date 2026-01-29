package handlers

import "github.com/labstack/echo/v4"

func jsonError(c echo.Context, code int, msg string) error {
	return c.JSON(code, map[string]string{"error": msg})
}
