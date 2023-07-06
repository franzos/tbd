package handler

import (
	"fmt"
	"strings"
	"tbd/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func modelPublicUser(u model.User) model.PublicUser {
	return model.PublicUser{
		ID:        u.ID,
		Data:      u.Data,
		CreatedAt: u.CreatedAt,
	}
}

func userFromToken(c echo.Context) (model.User, error) {
	jwtToken := c.Get("user").(*jwt.Token)

	claims, err := jwtToken.Claims.(*model.JwtCustomClaims)
	if !err {
		return model.User{}, fmt.Errorf("invalid token claims")
	}

	// To make sure it's a valid uuid
	id, pErr := uuid.Parse(claims.Subject)
	if pErr != nil {
		return model.User{}, fmt.Errorf("invalid subject; expected UUID: %v", pErr)
	}

	u := model.User{}
	u.ID = id.String()
	roles := strings.Split(claims.Roles, ",")
	u.Roles = roles

	return u, nil
}

func userIsAdmin(c echo.Context) bool {
	user, err := userFromToken(c)
	if err != nil {
		return false
	}

	for _, role := range user.Roles {
		if role == "admin" {
			return true
		}
	}

	return false
}
