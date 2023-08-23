package model

import (
	"github.com/biter777/countries"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"gorm.io/gorm"
)

// The slug is used on the local server; for ex. /city/berlin
// The glob_id is used to identifiy cities globally (all Berlin's in Germany will have the same glob_id de:berlin)
// That means there can be multiple de:berlin on a server, and in the world
type City struct {
	ID          string `json:"id" gorm:"type:uuid;primarykey"`
	Slug        string `json:"slug" gorm:"unique"`
	GlobID      string `json:"glob_id"`
	Name        string `json:"name" validate:"required"`
	CountryCode string `json:"country_code" validate:"required"`
	State       string `json:"state"`
}

// Slug is the URL under which the city is accessible
type PublicCity struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	GlobID      string `json:"glob_id"`
	Name        string `json:"name"`
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
	State       string `json:"state"`
}

func (base *City) BeforeCreate(tx *gorm.DB) (err error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	// This is flexible to accomodate partial input
	// Community maintainers can merge cities as needed
	base.Slug = CitySlug(base.Name)
	base.GlobID = CitySlugAuto(base.CountryCode, base.State, base.Name)

	base.ID = id.String()
	return
}

func CityStateCountrySlug(country, state, name string) string {
	return slug.Make(country) + ":" + slug.Make(state) + ":" + slug.Make(name)
}

func CityCountrySlug(country, name string) string {
	return slug.Make(country) + ":" + slug.Make(name)
}

func CitySlug(name string) string {
	return slug.Make(name)
}

func CitySlugAuto(country, state, name string) string {
	if state != "" && country != "" {
		return CityStateCountrySlug(country, state, name)
	} else if country != "" {
		return CityCountrySlug(country, name)
	} else {
		return CitySlug(name)
	}
}

func (c City) ToPublicFormat() PublicCity {
	pc := PublicCity{}
	country := countries.ByName(c.CountryCode)

	pc.ID = c.ID
	pc.Slug = c.Slug
	pc.GlobID = c.CountryCode + ":" + c.Slug
	pc.Name = c.Name
	pc.CountryCode = c.CountryCode
	// Default, to be removed once we have i18n
	pc.CountryName = country.Info().Name
	pc.State = c.State
	return pc
}
