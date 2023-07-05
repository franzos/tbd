package main

import (
	"os"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/subosito/gotenv"

	"tbd/handler"
	"tbd/model"
)

func main() {
	gotenv.Load()
	checkConfig()

	e := echo.New()
	e.Logger.SetLevel(log.ERROR)
	// e.Use(middleware.Logger())

	config := echojwt.Config{
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

	e.Use(echojwt.WithConfig(config))

	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            3600,
		ContentSecurityPolicy: "default-src 'self'",
	}))

	// Database connection and migration
	db, err := gorm.Open(sqlite.Open(DB_PATH()), &gorm.Config{})
	if err != nil {
		e.Logger.Fatal(err)
	}

	db.AutoMigrate(&model.User{}, &model.Entry{}, &model.File{})

	// Initialize handler
	h := &handler.Handler{DB: db}

	// Routes
	e.POST("/signup", h.Signup)
	e.POST("/login", h.Login)

	e.POST("/entries", h.CreateEntry)
	e.GET("/entries", h.FetchEntries)
	e.GET("/entries/:id", h.FetchEntry)
	e.PATCH("/entries/:id", h.UpdateEntry)
	e.DELETE("/entries/:id", h.DeleteEntry)

	e.POST("/files/multi", h.CreateFiles)

	e.GET("/account/me", h.Me)
	e.PATCH("/account/me", h.UpdateMe)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
