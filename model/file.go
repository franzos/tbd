package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type File struct {
	ID          string `json:"id" gorm:"type:uuid;primarykey"`
	Title       string `json:"title"`
	Path        string `json:"path"`
	Mime        string `json:"mime"`
	Size        int64  `json:"size"`
	CreatedByID string `json:"created_by_id"`
	CreatedBy   User   `json:"created_by" gorm:"foreignKey:CreatedByID"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   sql.NullTime `gorm:"index"`
}

func (base *File) BeforeCreate(tx *gorm.DB) (err error) {
	if base.ID != "" {
		return
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	base.ID = id.String()
	return
}
