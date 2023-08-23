package model

import (
	"fmt"
	"strings"
)

type Link struct {
	Url  string `json:"url"`
	Name string `json:"name"`
}

type ContactInfo struct {
	Medium   string `json:"medium"`
	Handle   string `json:"handle"`
	IsPublic bool   `json:"is_public"`
}

func (link *Link) ToJson() string {
	return fmt.Sprintf(`{"url": "%s", "name": "%s"}`, link.Url, link.Name)
}

type UserProfile struct {
	Links             []Link        `json:"links"`
	PublicContactInfo []ContactInfo `json:"public_contact_info"`
	Description       string        `json:"description"`
}

func (profile *UserProfile) IsEmpty() bool {
	return len(profile.Links) == 0 && profile.Description == ""
}

func (profile *UserProfile) ToJson() string {
	links := make([]string, len(profile.Links))
	for i, link := range profile.Links {
		links[i] = link.ToJson()
	}
	return fmt.Sprintf(`{"links": [%s], "description": "%s"}`, strings.Join(links, ","), profile.Description)
}
