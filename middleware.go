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
				return next(c)
			}
		}

		return echo.NewHTTPError(http.StatusForbidden, "unauthorized")
	}
}

func getJwtMVConfig() echojwt.Config {
	return echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(model.JwtCustomClaims)
		},
		SigningKey: []byte(os.Getenv("JWT_SECRET")),
		Skipper: func(c echo.Context) bool {
			/**
			Skip authentication for signup and login requests, as well as listing entries
			- login
			- signup
			- entries (list)
			- entries/:id (get)
			*/
			isEntriesList := c.Path() == "/entries" && c.Request().Method == "GET"
			isEntriesGet := c.Path() == "/entries/:id" && c.Request().Method == "GET"
			if c.Path() == "/login" || c.Path() == "/signup" || isEntriesList || isEntriesGet {
				return true
			}
			return false
		},
	}
}
