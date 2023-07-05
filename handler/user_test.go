package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/icrowley/fake"
	"github.com/stretchr/testify/assert"
)

func signupAndLogin(t *testing.T) string {
	signupURL := "http://localhost:1323/signup"
	signupData := map[string]string{
		"email":    fake.EmailAddress(),
		"password": "password123",
	}
	signupPayload, _ := json.Marshal(signupData)

	signupReq, _ := http.NewRequest(http.MethodPost, signupURL, bytes.NewBuffer(signupPayload))
	signupReq.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	signupRec, err := client.Do(signupReq)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusCreated, signupRec.StatusCode)

	loginURL := "http://localhost:1323/login"
	loginData := map[string]string{
		"email":    signupData["email"],
		"password": signupData["password"],
	}
	loginPayload, _ := json.Marshal(loginData)

	loginReq, _ := http.NewRequest(http.MethodPost, loginURL, bytes.NewBuffer(loginPayload))
	loginReq.Header.Set("Content-Type", "application/json")

	loginRec, err := client.Do(loginReq)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, loginRec.StatusCode)

	var loginResponse struct {
		Token string `json:"token"`
	}
	err = json.NewDecoder(loginRec.Body).Decode(&loginResponse)
	assert.NoError(t, err)

	return loginResponse.Token
}

func TestSignup(t *testing.T) {
	signupURL := "http://localhost:1323/signup"
	signupData := map[string]string{
		"email":    fake.EmailAddress(),
		"password": "password123",
	}
	signupPayload, _ := json.Marshal(signupData)

	signupReq, _ := http.NewRequest(http.MethodPost, signupURL, bytes.NewBuffer(signupPayload))
	signupReq.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	signupRec, err := client.Do(signupReq)
	assert.NoError(t, err)

	// Assertions for signup
	assert.Equal(t, http.StatusCreated, signupRec.StatusCode)
}

func TestLoginWithCorrectCredentials(t *testing.T) {
	// Signup
	signupURL := "http://localhost:1323/signup"
	signupData := map[string]string{
		"email":    fake.EmailAddress(),
		"password": "password123",
	}
	signupPayload, _ := json.Marshal(signupData)

	signupReq, _ := http.NewRequest(http.MethodPost, signupURL, bytes.NewBuffer(signupPayload))
	signupReq.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	_, err := client.Do(signupReq)
	assert.NoError(t, err)

	// Login
	loginURL := "http://localhost:1323/login"
	loginData := map[string]string{
		"email":    signupData["email"],
		"password": signupData["password"],
	}
	loginPayload, _ := json.Marshal(loginData)

	loginReq, _ := http.NewRequest(http.MethodPost, loginURL, bytes.NewBuffer(loginPayload))
	loginReq.Header.Set("Content-Type", "application/json")

	loginRec, err := client.Do(loginReq)
	assert.NoError(t, err)

	// Assertions for login with correct credentials
	assert.Equal(t, http.StatusOK, loginRec.StatusCode)
}

func TestLoginWithWrongEmail(t *testing.T) {
	// Login
	loginURL := "http://localhost:1323/login"
	loginData := map[string]string{
		"email":    "wrong@example.com",
		"password": "password123",
	}
	loginPayload, _ := json.Marshal(loginData)

	loginReq, _ := http.NewRequest(http.MethodPost, loginURL, bytes.NewBuffer(loginPayload))
	loginReq.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	loginRec, err := client.Do(loginReq)
	assert.NoError(t, err)

	// Assertions for login with wrong email
	assert.Equal(t, http.StatusUnauthorized, loginRec.StatusCode)
}

func TestLoginWithWrongPassword(t *testing.T) {
	// Signup
	signupURL := "http://localhost:1323/signup"
	signupData := map[string]string{
		"email":    fake.EmailAddress(),
		"password": "password123",
	}
	signupPayload, _ := json.Marshal(signupData)

	signupReq, _ := http.NewRequest(http.MethodPost, signupURL, bytes.NewBuffer(signupPayload))
	signupReq.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	_, err := client.Do(signupReq)
	assert.NoError(t, err)

	// Login
	loginURL := "http://localhost:1323/login"
	loginData := map[string]string{
		"email":    signupData["email"],
		"password": "wrongpassword",
	}
	loginPayload, _ := json.Marshal(loginData)

	loginReq, _ := http.NewRequest(http.MethodPost, loginURL, bytes.NewBuffer(loginPayload))
	loginReq.Header.Set("Content-Type", "application/json")

	loginRec, err := client.Do(loginReq)
	assert.NoError(t, err)

	// Assertions for login with wrong password
	assert.Equal(t, http.StatusUnauthorized, loginRec.StatusCode)
}

func TestAccountMe(t *testing.T) {
	token := signupAndLogin(t)

	// Use the token to request the user's own profile
	accountURL := "http://localhost:1323/account/me"
	accountReq, _ := http.NewRequest(http.MethodGet, accountURL, nil)
	accountReq.Header.Set("Authorization", "Bearer "+token)

	client := http.Client{}
	accountRec, err := client.Do(accountReq)
	assert.NoError(t, err)

	// Assertions for requesting user profile
	assert.Equal(t, http.StatusOK, accountRec.StatusCode)
}
