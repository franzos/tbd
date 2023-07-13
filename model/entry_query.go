package model

type EntryQueryParams struct {
	Offset    int    `query:"offset" validate:"omitempty,number,min=0"`
	Limit     int    `query:"limit" validate:"omitempty,number,min=1"`
	Type      string `query:"type"`
	Price     string `query:"price"`
	StartDate string `query:"start_date"`
	EndDate   string `query:"end_date"`
	Country   string `query:"country"`
	City      string `query:"city"`
}
