package handler

import (
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"gorm.io/gorm"

	"tbd/model"
)

func (h *Handler) Signup(c echo.Context) error {
	u := model.User{}
	if err := c.Bind(&u); err != nil {
		return err
	}

	if err := c.Validate(&u); err != nil {
		return err
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return &echo.HTTPError{Code: http.StatusInternalServerError}
	}
	u.Password = string(hash)

	r := h.DB.Create(&u)
	if r.Error != nil {
		// r.Error == gorm.ErrDuplicatedKey is not getting caught
		if (r.Error).Error() == "UNIQUE constraint failed: users.email" {
			return &echo.HTTPError{Code: http.StatusConflict, Message: "User with email already exists. Reset password?"}
		}
		return &echo.HTTPError{Code: http.StatusInternalServerError}
	}

	u.Password = ""
	return c.JSON(http.StatusCreated, u)
}

func (h *Handler) Login(c echo.Context) error {
	f := model.UserLogin{}
	if err := c.Bind(&f); err != nil {
		return err
	}

	if err := c.Validate(&f); err != nil {
		return err
	}

	u := model.User{}
	r := h.DB.Where("email = ?", f.Email).First(&u)
	if r.Error != nil {
		if r.Error == gorm.ErrRecordNotFound {
			return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "Invalid email or password."}
		}
		return &echo.HTTPError{Code: http.StatusInternalServerError}
	}

	// Check password hash
	valid := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(f.Password))
	if valid != nil {
		return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "Invalid email or password."}
	}

	// Assemble JWT
	claims := &model.JwtCustomClaims{
		Roles: "default",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
			Subject:   u.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Something went wrong. Please try again later."}
	}

	return c.JSON(http.StatusOK, model.UserLoginResponse{Token: signedToken})
}

func (h *Handler) Me(c echo.Context) error {
	u, httpErr := userFromContext(c)
	if httpErr != nil {
		// server error
		return httpErr
	}

	r := h.DB.First(&u)
	if r.Error != nil {
		if r.Error == gorm.ErrRecordNotFound {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "User not found. Please try again later."}
		}
		return &echo.HTTPError{Code: http.StatusInternalServerError}
	}

	u.Password = ""
	return c.JSON(http.StatusOK, u)
}

func (h *Handler) UpdateMe(c echo.Context) error {
	u, httpErr := userFromContext(c)
	if httpErr != nil {
		// server error
		return httpErr
	}

	nu := model.User{}
	if err := c.Bind(&nu); err != nil {
		return err
	}

	// We really only allow updating the data field for now
	r := h.DB.Model(model.User{ID: u.ID}).Update("data", nu.Data)
	if r.Error != nil {
		if r.Error == gorm.ErrRecordNotFound {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "User not found. Please try again later."}
		}
		return &echo.HTTPError{Code: http.StatusInternalServerError}
	}

	return c.JSON(http.StatusOK, UpdateResponse{Updated: r.RowsAffected})
}
