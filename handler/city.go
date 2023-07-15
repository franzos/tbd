package handler

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"tbd/model"

	"github.com/biter777/countries"
	"gorm.io/gorm"
)

func (h *Handler) GetAndCreateIfNotFoundCity(address model.Address) (*model.City, error) {
	var city model.City
	city.Name = address.City

	if address.Country != "" {
		country := countries.ByName(address.Country)
		unknown := countries.Unknown
		// If unknown, we simply ignore this
		if country.Alpha2() != unknown.Alpha3() {
			city.CountryCode = strings.ToLower(country.Alpha2())

			if address.State != "" {
				states := country.Subdivisions()
				// lowercase and trim whitespcace
				state := strings.ToLower(strings.TrimSpace(address.State))
				for _, s := range states {
					if state == strings.ToLower(strings.TrimSpace(s.String())) {
						city.State = s.String()
						break
					}
				}
			}
		}
	}

	slug := model.CitySlugAuto(city.CountryCode, city.State, city.Name)

	err := h.DB.Where("slug = ?", slug).First(&city).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Println(fmt.Sprintf("Creating city %s", slug))
			err = h.DB.Create(&city).Error
			if err != nil {
				return nil, err
			}
		} else {
			log.Println(fmt.Sprintf("Error getting city %s", slug))
			return nil, err
		}
	}

	return &city, nil
}
