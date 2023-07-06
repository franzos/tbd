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
func UserFromContext(c echo.Context) (model.User, error) {
	user := c.Get("user")
	if user == nil {
		return model.User{}, fmt.Errorf("user not found in context")
	}
	jwtToken := user.(*jwt.Token)

	claims, err := jwtToken.Claims.(*model.JwtCustomClaims)
	if !err {
		return model.User{}, fmt.Errorf("error getting claims")
	}

	// To make sure it's a valid uuid
	id, pErr := uuid.Parse(claims.Subject)
	if pErr != nil {
		return model.User{}, fmt.Errorf("error parsing uuid")
	}

	u := model.User{}
	u.ID = id.String()
	roles := strings.Split(claims.Roles, ",")
	u.Roles = roles

	return u, nil
}

func UserFromContextHttpError(c echo.Context) (model.User, *echo.HTTPError) {
	result, err := UserFromContext(c)
	if err != nil {
		return model.User{}, &echo.HTTPError{Code: 500, Message: err}
	}
	return result, nil
}
