package model

import (
	"regexp"
	"strings"
)

func StripUsername(username string) string {
	stripped := strings.ReplaceAll(username, " ", "")
	return strings.ToLower(stripped)
}

func StripEmail(email string) *string {
	stripped := strings.ReplaceAll(email, " ", "")
	stripped = strings.ToLower(stripped)
	return &stripped
}

func StripPhone(phone string) *string {
	stripped := strings.ReplaceAll(phone, "-", "")
	stripped = strings.ReplaceAll(stripped, " ", "")
	stripped = strings.ReplaceAll(stripped, "(", "")
	stripped = strings.ReplaceAll(stripped, ")", "")
	stripped = strings.ToLower(stripped)
	return &stripped
}

func IsValidUsername(username string) bool {
	length := len(username) >= 3 && len(username) <= 20
	chars := regexp.MustCompile(`^[\w\-._~]+$`).MatchString(username)
	return length && chars
}

func IsValidEmail(email string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`).MatchString(email)
}

func IsValidPhone(phone string) bool {
	return regexp.MustCompile(`^\+?[0-9]{10,15}$`).MatchString(phone)
}
