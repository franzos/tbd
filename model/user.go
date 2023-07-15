package model

import (
	"database/sql"
	"log"
	"os"
	"regexp"
	"strings"
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
	ImageID     *string        `json:"image_id"`
	Image       *File          `json:"image,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Name        string         `json:"name"`
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

func (base *User) BeforeCreate(tx *gorm.DB) (err error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	name := base.Name
	email := *base.Email
	passphrase := os.Getenv("PGP_PASSPHRASE")

	keyPair, err := pgp.GenerateKeyPair(name, email, passphrase)
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

func (user User) IsAdmin() bool {
	for _, v := range user.Roles {
		if v == "admin" {
			return true
		}
	}

	return false
}

func StripUsername(username string) string {
	stripped := strings.ReplaceAll(username, " ", "")
	return strings.ToLower(stripped)
}

func IsValidUsername(username string) bool {
	length := len(username) >= 3 && len(username) <= 20
	chars := regexp.MustCompile(`^[\w\-._~]+$`).MatchString(username)
	return length && chars
}

func UsernameWithLocalPart(username, domain string) string {
	return "@" + domain + ":" + username
}

func StripEmail(email string) *string {
	stripped := strings.ReplaceAll(email, " ", "")
	stripped = strings.ToLower(stripped)
	return &stripped
}

func IsValidEmail(email string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`).MatchString(email)
}

func StripPhone(phone string) *string {
	stripped := strings.ReplaceAll(phone, "-", "")
	stripped = strings.ReplaceAll(stripped, " ", "")
	stripped = strings.ReplaceAll(stripped, "(", "")
	stripped = strings.ReplaceAll(stripped, ")", "")
	stripped = strings.ToLower(stripped)
	return &stripped
}

func IsValidPhone(phone string) bool {
	return regexp.MustCompile(`^\+?[0-9]{10,15}$`).MatchString(phone)
}
