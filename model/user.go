package model

import (
	"database/sql"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Primary user struct for DB interactions
type User struct {
	ID        string         `json:"id" gorm:"type:uuid;primarykey"`
	Email     string         `json:"email" gorm:"uniqueIndex" validate:"required,email"`
	Phone     string         `json:"phone,omitempty"`
	Password  string         `json:"password,omitempty" validate:"required"`
	Roles     []string       `json:"roles" gorm:"serializer:json;default:'[]'"`
	Profile   UserProfile    `json:"profile" gorm:"serializer:json"`
	Data      datatypes.JSON `json:"data"`
	IsListed  bool           `json:"is_listed" gorm:"default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime `gorm:"index"`
}

// User extracted from JWT token
type AuthUser struct {
	ID      string   `json:"id"`
	Roles   []string `json:"roles"`
	IsAdmin bool     `json:"is_admin"`
}

// User to be returned to client
type PublicUser struct {
	ID        string      `json:"id"`
	Profile   UserProfile `json:"profile"`
	CreatedAt time.Time   `json:"created_at"`
}

func (base *User) BeforeCreate(tx *gorm.DB) (err error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	base.ID = id.String()
	return
}

type UserLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserLoginResponse struct {
	Token string `json:"token"`
}

type JwtCustomClaims struct {
	Roles string `json:"roles"`
	jwt.RegisteredClaims
}

func (user User) ToPublicFormat() interface{} {
	return PublicUser{
		ID:        user.ID,
		Profile:   user.Profile,
		CreatedAt: user.CreatedAt,
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
