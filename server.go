package main

import (
	"fmt"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/go-playground/validator"
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

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func main() {
	gotenv.Load()

	e := echo.New()
	e.Logger.SetLevel(log.ERROR)

	checkConfig()

	// Database connection and migration
	db, err := gorm.Open(sqlite.Open(DB_PATH()), &gorm.Config{})
	if err != nil {
		e.Logger.Fatal(err)
	}

	db.AutoMigrate(&model.User{}, &model.Entry{}, &model.File{})

	// e.Use(middleware.Logger())

	// Saniztize
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            3600,
		ContentSecurityPolicy: "default-src 'self'",
	}))

	// CORS default
	// Allows requests from any origin wth GET, HEAD, PUT, POST or DELETE method.
	e.Use(middleware.CORS())

	// CORS restricted
	// Allows requests from any `https://labstack.com` or `https://labstack.net` origin
	// wth GET, PUT, POST or DELETE method.
	// e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
	// 	AllowOrigins: []string{"https://labstack.com", "https://labstack.net"},
	// 	AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	// }))

	// Authenticate
	e.Use(echojwt.WithConfig(getJwtMVConfig()))

	// Authorize

	// authEnforcer
	authEnforcer, err := casbin.NewEnforcer("./auth_model.conf", "./policy.csv")
	if err != nil {
		panic(fmt.Sprintf("failed to create casbin enforcer: %s", err))
	}

	polErr := authEnforcer.LoadPolicy()
	if polErr != nil {
		panic(fmt.Sprintf("failed to load policy: %s", err))
	}

	e.Use(AuthorizationMW{Enforcer: authEnforcer}.Authorize)

	// Initialize handler
	e.Validator = &CustomValidator{validator: validator.New()}
	h := &handler.Handler{DB: db}

	// Routes
	e.POST("/signup", h.Signup)
	e.POST("/login", h.Login)

	e.GET("/users", h.FetchUsers)
	e.GET("/users/:id", h.FetchUser)
	e.DELETE("/users/:id", h.DeleteUser)

	e.POST("/entries", h.CreateEntry)
	e.GET("/entries", h.FetchEntries)
	e.GET("/entries/by-city/count", h.EntriesByCity)
	e.GET("/entries/by-country/count", h.EntriesByCountry)
	e.GET("/entries/by-type/count", h.EntriesByType)
	e.GET("/entries/:id", h.FetchEntry)
	e.PATCH("/entries/:id", h.UpdateEntry)
	e.DELETE("/entries/:id", h.DeleteEntry)

	e.GET("/files", h.FetchFiles)
	e.POST("/files/multi", h.CreateFiles)
	e.DELETE("/files/:id", h.DeleteFile)
	e.GET("/files/:id/download", h.DownloadFile)

	e.GET("/account/me", h.Me)
	e.PATCH("/account/me", h.UpdateMe)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
