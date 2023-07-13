package model

import (
	"github.com/biter777/countries"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"gorm.io/gorm"
)

type City struct {
	ID      string `json:"id" gorm:"type:uuid;primarykey"`
	Slug    string `json:"slug" gorm:"unique"`
	Name    string `json:"name" validate:"required"`
	Country string `json:"country" validate:"required"`
	State   string `json:"state"`
}

type PublicCity struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Country     string `json:"country"`
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
	base.Slug = CitySlugAuto(base.Country, base.State, base.Name)

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
	country := countries.ByName(c.Country)
	pc.ID = c.ID
	pc.Name = c.Name
	pc.Country = c.Country
	pc.CountryName = country.Info().Name
	pc.State = c.State
	return pc
}
