package model

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"tbd/pgp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Primary user struct for DB interactions
type User struct {
	ID          string         `json:"id" gorm:"type:uuid;primarykey"`
	ImageID     string         `json:"image_id"`
	Image       *File          `json:"image" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Name        *string        `json:"name"`
	Username    string         `json:"username" gorm:"uniqueIndex"`
	Email       *string        `json:"email" gorm:"uniqueIndex" validate:"email"`
	Phone       *string        `json:"phone,omitempty" gorm:"uniqueIndex"`
	Password    string         `json:"password,omitempty"`
	Roles       []string       `json:"roles" gorm:"serializer:json;default:'[]'"`
	Profile     UserProfile    `json:"profile" gorm:"serializer:json"`
	Data        datatypes.JSON `json:"data"`
	IsConfirmed bool           `json:"is_confirmed" gorm:"default:false"`
	IsListed    bool           `json:"is_listed" gorm:"default:false"`
	PrivateKey  string         `json:"private_key"`
	PublicKey   string         `json:"public_key"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   sql.NullTime `gorm:"index"`
}

// Signup a new user
// Email or Phone is required for verification and notification
type SignupUserReq struct {
	Name     string `json:"name"`
	Username string `json:"username" gorm:"uniqueIndex"`
	Email    string `json:"email" gorm:"uniqueIndex" validate:"email"`
	Phone    string `json:"phone,omitempty"`
	Password string `json:"password,omitempty" validate:"required"`
	IsListed bool   `json:"is_listed" gorm:"default:false"`
}

// Login an existing user
// The user may login with either username, email or phone number
// The type determines which option to use; default is email
// Supported types are: username, email, phone
type LoginUserReq struct {
	Type     string `json:"type"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password" validate:"required"`
}

// User extracted from JWT token
type AuthUser struct {
	ID      string   `json:"id"`
	Roles   []string `json:"roles"`
	IsAdmin bool     `json:"is_admin"`
}

// User to be returned to client
type PublicUser struct {
	ID                    string      `json:"id"`
	Image                 PublicFile  `json:"image,omitempty"`
	Username              string      `json:"username"`
	UsernameWithLocalPart string      `json:"username_with_local_part"`
	Profile               UserProfile `json:"profile"`
	PublicKey             string      `json:"public_key"`
	CreatedAt             time.Time   `json:"created_at"`
}

// User to be returned only on /me endpoint
type PrivateUser struct {
	ID                    string      `json:"id"`
	Name                  *string     `json:"name"`
	Email                 *string     `json:"email"`
	Phone                 *string     `json:"phone"`
	Image                 PublicFile  `json:"image,omitempty"`
	Username              string      `json:"username"`
	UsernameWithLocalPart string      `json:"username_with_local_part"`
	Profile               UserProfile `json:"profile"`
	PublicKey             string      `json:"public_key"`
	CreatedAt             time.Time   `json:"created_at"`
}

func (base *User) BeforeCreate(tx *gorm.DB) (err error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	name := base.Name
	email := *base.Email
	passphrase := os.Getenv("PGP_PASSPHRASE")

	keyPair, err := pgp.GenerateKeyPair(*name, email, passphrase)
	if err != nil {
		log.Println(err)
	} else {
		base.PrivateKey = keyPair.PrivateKey
		base.PublicKey = keyPair.PublicKey
	}

	base.ID = id.String()
	return
}

type LoginUserReqResponse struct {
	Token string `json:"token"`
}

type JwtCustomClaims struct {
	Roles string `json:"roles"`
	jwt.RegisteredClaims
}

func (user User) ToPublicFormat(domain string) interface{} {
	return PublicUser{
		ID:                    user.ID,
		Username:              user.Username,
		UsernameWithLocalPart: UsernameWithLocalPart(user.Username, domain),
		Profile:               user.Profile,
		PublicKey:             user.PublicKey,
		CreatedAt:             user.CreatedAt,
	}
}

func (user User) ToUserPrivateFormat(domain string) PrivateUser {
	return PrivateUser{
		ID:                    user.ID,
		Name:                  user.Name,
		Email:                 user.Email,
		Phone:                 user.Phone,
		Username:              user.Username,
		UsernameWithLocalPart: UsernameWithLocalPart(user.Username, domain),
		Profile:               user.Profile,
		PublicKey:             user.PublicKey,
		CreatedAt:             user.CreatedAt,
	}
}

func (user User) IsAdmin() bool {
	for _, v := range user.Roles {
		if v == "admin" {
			return true
		}
	}

	return false
}

func UsernameWithLocalPart(username, domain string) string {
	return "@" + domain + ":" + username
}

func (user SignupUserReq) Strip() {
	if user.Username != "" {
		user.Username = StripUsername(user.Username)
	}
	if user.Email != "" {
		user.Email = *StripEmail(user.Email)
	}
	if user.Phone != "" {
		user.Phone = *StripPhone(user.Phone)
	}
}

func (user SignupUserReq) Validate() error {
	if user.Username != "" {
		if !IsValidUsername(user.Username) {
			return errors.New("Invalid username")
		}
	}
	if user.Email != "" {
		if !IsValidEmail(user.Email) {
			return errors.New("Invalid email")
		}
	}
	if user.Phone != "" {
		if !IsValidPhone(user.Phone) {
			return errors.New("Invalid phone")
		}
	}
	if user.Email == "" && user.Phone == "" {
		return errors.New("Email or phone is required")
	}

	return nil
}
