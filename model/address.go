package model

type Address struct {
	Building  string `json:"building"`
	Street    string `json:"street"`
	Number    string `json:"number"`
	PostCode  string `json:"post_code"`
	City      string `json:"city"`
	State     string `json:"state"`
	Country   string `json:"country"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}
