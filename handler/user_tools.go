package handler

import (
	"strings"
	"tbd/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// This returns a HTTP errror if anything goes wrong; to be used directly in the handler
func userFromContext(c echo.Context) (model.User, *echo.HTTPError) {
	jwtToken := c.Get("user").(*jwt.Token)

	claims, err := jwtToken.Claims.(*model.JwtCustomClaims)
	if !err {
		return model.User{}, &echo.HTTPError{Code: 500, Message: "error parsing claims"}
	}

	// To make sure it's a valid uuid
	id, pErr := uuid.Parse(claims.Subject)
	if pErr != nil {
		return model.User{}, &echo.HTTPError{Code: 500, Message: "invalid subject; expected UUID"}
	}

	u := model.User{}
	u.ID = id.String()
	roles := strings.Split(claims.Roles, ",")
	u.Roles = roles

	return u, nil
}
