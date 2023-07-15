package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jaswdr/faker"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"gorm.io/gorm"

	"tbd/model"
)

func (h *Handler) FetchUsers(c echo.Context) error {
	offset, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	// Defaults
	if offset == 0 {
		offset = 1
	}
	if limit == 0 {
		limit = 100
	}

	count := int64(0)
	users := []model.User{}
	r := h.DB.Model(&model.User{}).
		Order("created_at desc").
		Count(&count).
		Preload("Image").
		Offset((offset - 1) * limit).
		Limit(limit).
		Find(&users)
	if r.Error != nil {
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to fetch users."}
	}

	return c.JSON(http.StatusOK, ListResponse{Total: int64(count), Items: responseArrFormatter[model.User](users, nil, os.Getenv("DOMAIN"))})
}

func (h *Handler) FetchUser(c echo.Context) error {
	id := c.Param("id")

	var user = model.User{ID: id}

	r := h.DB.First(&user).Preload("Image")
	if r.Error != nil {
		if r.Error == gorm.ErrRecordNotFound {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "User not found."}
		}
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to fetch user."}
	}

	return c.JSON(http.StatusOK, responseFormatter[model.User](user, nil, os.Getenv("DOMAIN")))
}

func (h *Handler) DeleteUser(c echo.Context) error {
	err := isSelfOrAdmin(c, c.Param("id"))
	if err != nil {
		return err
	}

	id := c.Param("id")

	var user = model.User{ID: id}

	r := h.DB.Delete(&user)
	if r.Error != nil {
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to delete user."}
	}

	// TODO: Delete user's entries, files, etc.

	return c.NoContent(http.StatusOK)
}

func usernameFromSignup(u model.SignupUserReq, tryCount int) (string, error) {
	username := ""

	if u.Username != "" {
		if !model.IsValidUsername(u.Username) {
			return "", &echo.HTTPError{Code: http.StatusBadRequest, Message: "Username can only contain letters, numbers, and other url-safe characters."}
		}
		username = u.Username
	} else if u.Name != "" {
		// Merge spaces with dot, lowercase, and trim
		username = strings.ToLower(strings.ReplaceAll(u.Name, " ", "."))
	} else {
		username = faker.New().Internet().User()
	}

	if tryCount > 0 {
		// add random number 0-100 to username
		username = username + strconv.Itoa(rand.Intn(100))
	}

	return username, nil
}

func (h *Handler) Signup(c echo.Context) error {
	u := model.SignupUserReq{}
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
	// TODO: Set unconfirmed until email or phone is confirmed

	newUser := model.User{
		Roles:    []string{"member"},
		Password: string(hash),
	}

	if u.Name != "" {
		newUser.Name = u.Name
	}

	if u.Email != "" && u.Phone != "" {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Email or phone is required to signup."}
	}

	if u.Email != "" {
		newUser.Email = model.StripEmail(u.Email)
		if !model.IsValidEmail(*newUser.Email) {
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid email."}
		}
	} else {
		newUser.Email = nil
	}

	if u.Phone != "" {
		newUser.Phone = model.StripPhone(u.Phone)
		if !model.IsValidPhone(*newUser.Phone) {
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid phone number."}
		}
	} else {
		newUser.Phone = nil
	}

	usernameIsUnique := false
	tryCount := 0
	for !usernameIsUnique {
		username, err := usernameFromSignup(u, tryCount)
		if err != nil {
			return err
		}
		// Check DB if username is unique
		var existingUser model.User
		r := h.DB.Where("username = ?", username).First(&existingUser)
		if r.Error != nil {
			if r.Error == gorm.ErrRecordNotFound {
				log.Println("NOT FOUND - username is unique")
				usernameIsUnique = true
				newUser.Username = username
			} else {
				log.Println("ERROR")
				return &echo.HTTPError{Code: http.StatusInternalServerError}
			}
		}
		tryCount++
		log.Println("tryCount", tryCount)
	}

	r := h.DB.Create(&newUser)
	if r.Error != nil {
		// r.Error == gorm.ErrDuplicatedKey is not getting caught
		if (r.Error).Error() == "UNIQUE constraint failed: users.username" {
			log.Println(fmt.Sprintf("Username %s already exists", u.Username))
			return &echo.HTTPError{Code: http.StatusConflict, Message: "User already exists. Reset password?"}
		}
		if (r.Error).Error() == "UNIQUE constraint failed: users.email" {
			log.Println(fmt.Sprintf("Email %s already exists", u.Email))
			return &echo.HTTPError{Code: http.StatusConflict, Message: "User already exists. Reset password?"}
		}
		if (r.Error).Error() == "UNIQUE constraint failed: users.phone" {
			log.Println(fmt.Sprintf("Phone %s already exists", u.Phone))
			return &echo.HTTPError{Code: http.StatusConflict, Message: "User already exists. Reset password?"}
		}

		return &echo.HTTPError{Code: http.StatusInternalServerError}
	}

	u.Password = ""
	return c.JSON(http.StatusCreated, u)
}

func (h *Handler) Login(c echo.Context) error {
	f := model.LoginUserReq{}
	if err := c.Bind(&f); err != nil {
		return err
	}

	if err := c.Validate(&f); err != nil {
		return err
	}

	// username, email, or phone
	loginType := ""
	if f.Type != "" {
		loginType = f.Type
	} else if f.Username != "" {
		loginType = "username"
		if !model.IsValidUsername(f.Username) {
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Username can only contain letters, numbers, and other url-safe characters."}
		}
	} else if f.Email != "" {
		loginType = "email"
		if !model.IsValidEmail(f.Email) {
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Improperly formatted email address."}
		}
	} else if f.Phone != "" {
		loginType = "phone"
		if !model.IsValidPhone(*model.StripPhone(f.Phone)) {
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Improperly formatted phone number."}
		}
	}

	// TODO: Check if user IsConfirmed
	u := model.User{}

	query := h.DB.Where("username = ?", model.StripUsername(f.Username))
	if loginType == "email" {
		query = h.DB.Where("email = ?", model.StripEmail(f.Email))
	} else if loginType == "phone" {
		query = h.DB.Where("phone = ?", model.StripPhone(f.Phone))
	}

	r := query.First(&u)
	if r.Error != nil {
		if r.Error != nil {
			if r.Error == gorm.ErrRecordNotFound {
				return &echo.HTTPError{Code: http.StatusUnauthorized, Message: fmt.Sprintf("Invalid %s or password.", loginType)}
			}
			return &echo.HTTPError{Code: http.StatusInternalServerError}
		}
	}

	// Check password hash
	valid := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(f.Password))
	if valid != nil {
		return &echo.HTTPError{Code: http.StatusUnauthorized, Message: fmt.Sprintf("Invalid %s or password.", loginType)}
	}

	// Assemble JWT
	claims := &model.JwtCustomClaims{
		Roles: strings.Join(u.Roles, ","),
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

	return c.JSON(http.StatusOK, model.LoginUserReqResponse{Token: signedToken})
}

func (h *Handler) Me(c echo.Context) error {
	reqUser := c.Get("user").(*model.AuthUser)

	u := model.User{ID: reqUser.ID}
	r := h.DB.First(&u)
	if r.Error != nil {
		if r.Error == gorm.ErrRecordNotFound {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "User not found. Please try again later."}
		}
		return &echo.HTTPError{Code: http.StatusInternalServerError}
	}

	u.Password = ""
	u.PrivateKey = ""
	return c.JSON(http.StatusOK, u)
}

func (h *Handler) UpdateMe(c echo.Context) error {
	reqUser := c.Get("user").(*model.AuthUser)

	nu := model.User{}
	if err := c.Bind(&nu); err != nil {
		return err
	}

	// We really only allow updating the data field for now
	if !nu.Profile.IsEmpty() {
		profileJSON, err := json.Marshal(nu.Profile)
		if err != nil {
			// handle error
			return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Error marshaling profile data"}
		}

		r := h.DB.Model(&model.User{ID: reqUser.ID}).Update("profile", string(profileJSON))
		if r.Error != nil {
			if r.Error == gorm.ErrRecordNotFound {
				return &echo.HTTPError{Code: http.StatusNotFound, Message: "User not found. Please try again later."}
			}
			return &echo.HTTPError{Code: http.StatusInternalServerError}
		}
	}

	if nu.Image != nil {
		current := model.User{ID: reqUser.ID}
		r := h.DB.First(&current).Preload("Image")
		if r.Error != nil {
			if r.Error == gorm.ErrRecordNotFound {
				return &echo.HTTPError{Code: http.StatusNotFound, Message: "User not found. Please try again later."}
			}
		}
		// TODO: Delete current image
		h.DB.Model(&r).Association("Image").Replace(&nu.Image)
	}

	return c.JSON(http.StatusOK, UpdateResponse{Updated: 1})
}
