package model

// TODO: Maybe call description 'body' instead?
type BaseEntry struct {
	Title       string  `json:"title" validate:"required"`
	Description string  `json:"description" validate:"required"`
	Address     Address `json:"address" validate:"required"`
}

type EntryApartmentShortTermRental struct {
	BaseEntry
	StartDate     string `json:"from" validate:"required"`
	EndDate       string `json:"to" validate:"required"`
	Price         string `json:"price" validate:"required"`
	PriceInterval string `json:"price_interval" validate:"required"`
}

type EntryApartmentLongTermRental struct {
	BaseEntry
	StartDate     string `json:"from" validate:"required"`
	Price         string `json:"price" validate:"required"`
	PriceInterval string `json:"price_interval" validate:"required"`
}

type EntryPetSitter struct {
	BaseEntry
	Price         string `json:"price" validate:"required"`
	PriceInterval string `json:"price_interval" validate:"required"`
}

type EntryItemSale struct {
	BaseEntry
	Price string `json:"price" validate:"required"`
}

type EntryLookingFor struct {
	BaseEntry
}
