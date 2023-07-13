package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var entryTypes = []string{
	// > 3 months
	"apartment-short-term-rental",
	// < 3 months
	"apartment-long-term-rental",
	"pet-sitter",
	"item-sale",
	"looking-for",
}

// Primary entry struct for DB interactions
// Data is signed and what's transferred should the user move communities
// City is extracted from Data and matched to the most applicable on the community; if none is found, one is created
// Flow is data -> entry; so if the user updates the City in entry.data.address.city, entry.city is updated
// Country is ISO code
// State is english name
type Entry struct {
	ID            string         `json:"id" gorm:"type:uuid;primarykey"`
	Type          string         `json:"type" validate:"required"`
	Data          datatypes.JSON `json:"data" validate:"required" gorm:"serializer:json"`
	DataSignature string         `json:"data_signature"`
	Files         []File         `json:"files,omitempty" gorm:"many2many:entry_files;"`
	CreatedByID   string         `json:"-"  gorm:"type:uuid"`
	CreatedBy     *User          `json:"created_by,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CityID        string         `json:"-" gorm:"type:uuid"`
	City          *City          `json:"city,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ExpiresAt     time.Time
}

// Entry to be returned to client
type PublicEntry struct {
	ID              string         `json:"id"`
	IDWithLocalPart string         `json:"id_with_local_part"`
	Type            string         `json:"type"`
	Data            datatypes.JSON `json:"data"`
	DataSignature   string         `json:"data_signature"`
	Files           []PublicFile   `json:"files,omitempty"`
	City            PublicCity     `json:"city,omitempty"`
	CreatedBy       PublicUser     `json:"created_by,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
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

func (e Entry) TypeIsValid() bool {
	for _, v := range entryTypes {
		if v == e.Type {
			return true
		}
	}
	return false
}

func (e Entry) ToPublicFormat(domain string) interface{} {
	pe := PublicEntry{}
	pe.ID = e.ID
	pe.IDWithLocalPart = EntryWithLocalPart(e.ID, domain)
	pe.Type = e.Type
	pe.Data = e.Data

	if e.DataSignature != "" {
		pe.DataSignature = e.DataSignature
	}

	if e.Files != nil {
		pe.Files = publicFilesFromFiles(e.Files)
	}

	if e.City != nil {
		pe.City = e.City.ToPublicFormat()
	}

	if e.CreatedBy != nil {
		pe.CreatedBy = e.CreatedBy.ToPublicFormat(domain).(PublicUser)
	}

	pe.CreatedAt = e.CreatedAt
	pe.UpdatedAt = e.UpdatedAt

	return pe
}

func EntryWithLocalPart(entryID, domain string) string {
	return "@" + domain + ":" + entryID
}
