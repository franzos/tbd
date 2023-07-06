package model

import (
	"database/sql"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type User struct {
	ID        string         `json:"id" gorm:"type:uuid;primarykey"`
	Email     string         `json:"email" gorm:"uniqueIndex" validate:"required,email"`
	Phone     string         `json:"phone,omitempty"`
	Password  string         `json:"password,omitempty" validate:"required"`
	Roles     []string       `json:"roles,omitempty" gorm:"type:text[]"`
	Data      datatypes.JSON `json:"data"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime `gorm:"index"`
}

type PublicUser struct {
	ID        string         `json:"id"`
	Data      datatypes.JSON `json:"data"`
	CreatedAt time.Time      `json:"created_at"`
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
