package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Entry struct {
	ID          string         `json:"id" gorm:"type:uuid;primarykey"`
	Type        string         `json:"type" validate:"required"`
	Data        datatypes.JSON `json:"data" validate:"required" gorm:"serializer:json"`
	Files       []File         `json:"files" gorm:"many2many:entry_files;"`
	CreatedByID string         `json:"-"`
	CreatedBy   User           `json:"created_by" gorm:"foreignKey:CreatedByID"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ExpiresAt   time.Time
	DeletedAt   sql.NullTime `gorm:"index"`
}

type PublicEntry struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Data      datatypes.JSON `json:"data"`
	Files     []PublicFile   `json:"files"`
	CreatedBy PublicUser     `json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
}

func (base *Entry) BeforeCreate(tx *gorm.DB) (err error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	base.ID = id.String()
	return
}

type SubmitEntry struct {
	Type  string         `json:"type" validate:"required"`
	Data  datatypes.JSON `json:"data" validate:"required"`
	Files []File         `json:"files" gorm:"many2many:entry_files;"`
}

var entryTypes = []string{
	// > 3 months
	"apartment-short-term-rental",
	// < 3 months
	"apartment-long-term-rental",
	"apartment-sale",
}

func (e Entry) TypeIsValid() bool {
	for _, v := range entryTypes {
		if v == e.Type {
			return true
		}
	}
	return false
}

func (e Entry) ToPublicFormat() interface{} {
	return PublicEntry{
		ID:        e.ID,
		Type:      e.Type,
		Data:      e.Data,
		Files:     publicFilesFromFiles(e.Files),
		CreatedBy: e.CreatedBy.ToPublicFormat(),
		CreatedAt: e.CreatedAt,
	}
}
