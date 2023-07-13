package main

import (
	"net/http"
	"os"
	"tbd/handler"
	"tbd/model"

	"github.com/casbin/casbin/v2"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

type AuthorizationMW struct {
	Enforcer *casbin.Enforcer
}

func (cfg AuthorizationMW) Authorize(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userRoles := []string{"anonymous"}
		user, err := handler.UserFromContext(c)
		if err == nil {
			userRoles = user.Roles
		}

		res := []bool{}

		for _, role := range userRoles {
			r, casbinErr := cfg.Enforcer.Enforce(role, c.Path(), c.Request().Method)
			if casbinErr != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "authorization erro")
			}
			res = append(res, r)
		}

		for _, r := range res {
			if r {
				c.Set("user", &user)
				return next(c)
			}
		}

		return echo.NewHTTPError(http.StatusForbidden, "unauthorized")
	}
}

type PublicPaths struct {
	Path   string
	Method string
}

var publicPaths = []PublicPaths{
	{
		Path:   "/login",
		Method: "POST",
	},
	{
		Path:   "/signup",
		Method: "POST",
	},
	{
		Path:   "/entries",
		Method: "GET",
	},
	{
		Path:   "/entries/:id",
		Method: "GET",
	},
	{
		Path:   "/files/:id/download",
		Method: "GET",
	},
	{
		Path:   "/entries/by-city/count",
		Method: "GET",
	},
	{
		Path:   "/entries/by-country/count",
		Method: "GET",
	},
	{
		Path:   "/entries/by-type/count",
		Method: "GET",
	},
}

func getJwtMVConfig() echojwt.Config {
	return echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(model.JwtCustomClaims)
		},
		ContextKey: "user_auth",
		SigningKey: []byte(os.Getenv("JWT_SECRET")),
		Skipper: func(c echo.Context) bool {
			for _, p := range publicPaths {
				if c.Path() == p.Path && c.Request().Method == p.Method {
					return true
				}
			}
			return false
		},
	}
}
