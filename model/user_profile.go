package model

type Link struct {
	Url  string `json:"url"`
	Name string `json:"name"`
}

type UserProfile struct {
	Links []Link `json:"links"`
}
