package handler

import (
	"fmt"
	"strings"
	"tbd/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// This returns a HTTP errror if anything goes wrong; to be used directly in the handler
func UserFromContext(c echo.Context) (model.AuthUser, error) {
	user := c.Get("user_auth")
	if user == nil {
		return model.AuthUser{}, fmt.Errorf("user not found in context")
	}
	jwtToken := user.(*jwt.Token)

	claims, err := jwtToken.Claims.(*model.JwtCustomClaims)
	if !err {
		return model.AuthUser{}, fmt.Errorf("error getting claims")
	}

	// To make sure it's a valid uuid
	id, pErr := uuid.Parse(claims.Subject)
	if pErr != nil {
		return model.AuthUser{}, fmt.Errorf("error parsing uuid")
	}

	roles := strings.Split(claims.Roles, ",")
	isAdmin := false
	for _, role := range roles {
		if role == "admin" {
			isAdmin = true
		}
	}

	u := model.AuthUser{
		ID:      id.String(),
		Roles:   strings.Split(claims.Roles, ","),
		IsAdmin: isAdmin,
	}

	return u, nil
}

func UserFromContextHttpError(c echo.Context) (model.AuthUser, *echo.HTTPError) {
	result, err := UserFromContext(c)
	if err != nil {
		return model.AuthUser{}, &echo.HTTPError{Code: 500, Message: err}
	}
	return result, nil
}
