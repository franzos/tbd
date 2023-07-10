package model

type EntryApartmentShortTermRental struct {
	Title         string  `json:"title" validate:"required"`
	Description   string  `json:"description" validate:"required"`
	StartDate     string  `json:"from" validate:"required"`
	EndDate       string  `json:"to" validate:"required"`
	Price         string  `json:"price" validate:"required"`
	PriceInterval string  `json:"price_interval" validate:"required"`
	Address       Address `json:"address" validate:"required"`
}

type EntryApartmentLongTermRental struct {
	Title         string  `json:"title" validate:"required"`
	Description   string  `json:"description" validate:"required"`
	StartDate     string  `json:"from" validate:"required"`
	Price         string  `json:"price" validate:"required"`
	PriceInterval string  `json:"price_interval" validate:"required"`
	Address       Address `json:"address" validate:"required"`
}
