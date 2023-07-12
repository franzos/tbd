package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Primary file struct for DB interactions
//
// Notes:
//   - IsProvisional indicates whether the file has ever been associated after uploaded
//     Files that have not been associated after X hours will be deleted
type File struct {
	ID            string `json:"id" gorm:"type:uuid;primarykey"`
	Title         string `json:"title"`
	Path          string `json:"path"`
	Mime          string `json:"mime"`
	Size          int64  `json:"size"`
	CreatedByID   string `json:"created_by_id"`
	CreatedBy     *User  `json:"created_by,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ExpiresAt     time.Time
	IsProvisional bool         `json:"is_provisional"`
	DeletedAt     sql.NullTime `gorm:"index"`
}

// File to be returned to client
type PublicFile struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Path  string `json:"path"`
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

func (f File) ToPublicFormat(domain string) interface{} {
	return PublicFile{
		ID:    f.ID,
		Title: f.Title,
		Path:  f.Path,
	}
}

func publicFilesFromFiles(files []File) []PublicFile {
	var publicFiles []PublicFile
	for _, v := range files {
		publicFiles = append(publicFiles, PublicFile{
			ID:    v.ID,
			Title: v.Title,
			Path:  v.Path,
		})
	}
	return publicFiles
}
