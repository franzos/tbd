package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TODO: This needs some consideration
// InResponseToID string   `json:"-"  gorm:"type:uuid"`
// InResponseTo   *Comment `json:"in_response_to,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
type Comment struct {
	ID          string `json:"id" gorm:"type:uuid;primarykey"`
	Body        string `json:"body" validate:"required"`
	EntryID     string `json:"-"  gorm:"type:uuid"`
	Entry       *Entry `json:"entry,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedByID string `json:"-"  gorm:"type:uuid"`
	CreatedBy   *User  `json:"created_by,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	DeletedAt   string `json:"deleted_at"`
}

// InResponseTo *PublicComment `json:"in_response_to,omitempty"`
type PublicComment struct {
	ID        string     `json:"id"`
	Body      string     `json:"body"`
	CreatedBy PublicUser `json:"created_by,omitempty"`
}

type MakeComment struct {
	EntryID string `json:"entry_id" validate:"required"`
	Body    string `json:"body" validate:"required"`
}

type EditComment struct {
	Body string `json:"body" validate:"required"`
}

func (c Comment) ToPublicFormat(domain string) interface{} {
	pc := PublicComment{
		ID:   c.ID,
		Body: c.Body,
	}

	if c.CreatedBy != nil {
		pc.CreatedBy = c.CreatedBy.ToPublicFormat(domain).(PublicUser)
	}

	return pc
}

func (base *Comment) BeforeCreate(tx *gorm.DB) (err error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	base.ID = id.String()
	return
}
