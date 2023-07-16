package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// To is either 0 or 1; 0 is a vote for the entry, 1 is a vote against the entry
type Vote struct {
	ID          string   `json:"id" gorm:"type:uuid;primarykey"`
	Vote        int      `json:"vote" validate:"required"`
	CreatedByID string   `json:"-"  gorm:"type:uuid"`
	CreatedBy   *User    `json:"created_by,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	EntryID     string   `json:"-"  gorm:"type:uuid"`
	Entry       *Entry   `json:"entry,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CommentID   string   `json:"-"  gorm:"type:uuid"`
	Comment     *Comment `json:"comment,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime `gorm:"index"`
}

type PublicVote struct {
	ID        string     `json:"id"`
	Vote      int        `json:"vote"`
	CreatedBy PublicUser `json:"created_by,omitempty"`
	CreatedAt time.Time
}

type CastVote struct {
	EntryID   string `json:"entry_id" validate:"required"`
	CommentID string `json:"comment_id" validate:"required"`
	Vote      int    `json:"vote" validate:"required"`
}

func (v Vote) ToPublicFormat(domain string) PublicVote {
	pv := PublicVote{
		ID:        v.ID,
		Vote:      v.Vote,
		CreatedAt: v.CreatedAt,
	}

	if v.CreatedBy != nil {
		pv.CreatedBy = v.CreatedBy.ToPublicFormat(domain).(PublicUser)
	}

	return pv
}

func (base *Vote) BeforeCreate(tx *gorm.DB) (err error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	base.ID = id.String()
	return
}
