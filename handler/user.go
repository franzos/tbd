package handler

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"gorm.io/gorm"

	"tbd/model"
)

func (h *Handler) Signup(c echo.Context) (err error) {
	u := model.User{}
	if err = c.Bind(&u); err != nil {
		return
	}

	// Validate
	if u.Email == "" || u.Password == "" {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid email or password."}
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
		return
	}

	u.Password = ""
	return c.JSON(http.StatusCreated, u)
}

func (h *Handler) Login(c echo.Context) (err error) {
	f := model.UserLogin{}
	if err = c.Bind(&f); err != nil {
		return
	}

	// Validate
	if f.Email == "" || f.Password == "" {
		return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "Invalid email or password."}
	}

	u := model.User{}
	r := h.DB.Where("email = ?", f.Email).First(&u)
	if r.Error != nil {
		if r.Error == gorm.ErrRecordNotFound {
			return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "Invalid email or password."}
		}
		return
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

func (h *Handler) Me(c echo.Context) (err error) {
	u, err := userFromToken(c)
	if err != nil {
		// server error
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: err}
	}

	r := h.DB.First(&u)
	if r.Error != nil {
		if r.Error == gorm.ErrRecordNotFound {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "User not found. Please try again later."}
		}
		return
	}

	u.Password = ""
	return c.JSON(http.StatusOK, u)
}

func (h *Handler) UpdateMe(c echo.Context) (err error) {
	u, err := userFromToken(c)
	if err != nil {
		// server error
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: err}
	}

	nu := model.User{}
	if err = c.Bind(&nu); err != nil {
		return
	}

	// We really only allow updating the data field for now
	r := h.DB.Model(model.User{ID: u.ID}).Update("data", nu.Data)
	if r.Error != nil {
		if r.Error == gorm.ErrRecordNotFound {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "User not found. Please try again later."}
		}
		return
	}

	return c.JSON(http.StatusOK, UpdateResponse{Updated: r.RowsAffected})
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

	// if err {
	// 	return 0, fmt.Errorf("user ID claim not found")
	// }

	return u, nil
}
