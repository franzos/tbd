package handler

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/icrowley/fake"
	"github.com/stretchr/testify/assert"

	"tbd/model"
)

func signupAndLogin(t *testing.T) string {
	signupURL := "http://localhost:1323/signup"
	signupData := model.SignupUserReq{
		Name:     fake.FullName(),
		Email:    fake.EmailAddress(),
		Password: "password123",
	}
	signupPayload, _ := json.Marshal(signupData)

	signupReq, _ := http.NewRequest(http.MethodPost, signupURL, bytes.NewBuffer(signupPayload))
	signupReq.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	signupRec, err := client.Do(signupReq)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusCreated, signupRec.StatusCode)

	loginURL := "http://localhost:1323/login"
	loginData := model.LoginUserReq{
		Email:    signupData.Email,
		Password: signupData.Password,
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

func updateUser(t *testing.T, token string, userData map[string]interface{}) {
	userURL := "http://localhost:1323/account/me"

	rec := performRequest(t, http.MethodPatch, userURL, token, userData)

	if rec.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK, got: %v", rec.StatusCode)
	}
}

func updateUserImage(t *testing.T, token string, image model.File) {
	updateURL := "http://localhost:1323/account/me"
	updateData := map[string]interface{}{
		"image": image,
	}

	payload, _ := json.Marshal(updateData)

	req, err := http.NewRequest(http.MethodPatch, updateURL, bytes.NewBuffer(payload))
	assert.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := http.Client{}
	res, err := client.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)
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
	signupData := model.SignupUserReq{
		Email:    fake.EmailAddress(),
		Password: "password123",
	}
	signupPayload, _ := json.Marshal(signupData)

	signupReq, _ := http.NewRequest(http.MethodPost, signupURL, bytes.NewBuffer(signupPayload))
	signupReq.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	_, err := client.Do(signupReq)
	assert.NoError(t, err)

	// Login
	loginURL := "http://localhost:1323/login"
	loginData := model.LoginUserReq{
		Email:    signupData.Email,
		Password: signupData.Password,
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
	loginData := model.LoginUserReq{
		Email:    "wrong@example.com",
		Password: "password123",
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
	signupData := model.SignupUserReq{
		Email:    fake.EmailAddress(),
		Password: "password123",
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
		"email":    signupData.Email,
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

func TestAccountMeUpdate(t *testing.T) {
	token := signupAndLogin(t)

	// Create a patch to update the profile
	updateData := map[string]interface{}{
		"profile": map[string]interface{}{
			"links": []map[string]interface{}{
				{
					"url":  "https://twitter.com/user",
					"name": "Twitter",
				},
			},
		},
	}

	// Update the user profile
	updateUser(t, token, updateData)

	// Use the token to request the user's own profile
	accountURL := "http://localhost:1323/account/me"
	accountReq, _ := http.NewRequest(http.MethodGet, accountURL, nil)
	accountReq.Header.Set("Authorization", "Bearer "+token)

	client := http.Client{}
	accountRec, err := client.Do(accountReq)
	assert.NoError(t, err)

	// Assertions for requesting user profile
	assert.Equal(t, http.StatusOK, accountRec.StatusCode)

	// Read the response body
	body, _ := ioutil.ReadAll(accountRec.Body)
	defer accountRec.Body.Close()

	// Parse the response into a map
	data := make(map[string]interface{})
	json.Unmarshal(body, &data)

	// Extract the profile from the response
	profile := data["profile"].(map[string]interface{})

	// Check the profile has been updated
	expectedLinks := []interface{}{
		map[string]interface{}{
			"url":  "https://twitter.com/user",
			"name": "Twitter",
		},
	}
	assert.Equal(t, expectedLinks, profile["links"])
}

func TestAccountMeUpdateImage(t *testing.T) {
	token := signupAndLogin(t)

	files := uploadFiles(t, token, "../concorde.jpg")

	// Update user image
	updateUserImage(t, token, files[0])

	// Get the updated user profile
	accountURL := "http://localhost:1323/account/me"
	accountReq, _ := http.NewRequest(http.MethodGet, accountURL, nil)
	accountReq.Header.Set("Authorization", "Bearer "+token)

	client := http.Client{}
	accountRec, err := client.Do(accountReq)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, accountRec.StatusCode)

	accountResponse := model.User{}

	err = json.NewDecoder(accountRec.Body).Decode(&accountResponse)
	assert.NoError(t, err)

	// Check if the image is correctly updated
	if accountResponse.Image != nil {
		assert.Equal(t, files[0].ID, accountResponse.Image.ID)
	} else {
		t.Fatal("Image is nil")
	}
}

func TestUserDeleteSelf(t *testing.T) {
	token := signupAndLogin(t)

	// Use the token to request the user's own profile
	accountURL := "http://localhost:1323/account/me"
	accountReq, _ := http.NewRequest(http.MethodGet, accountURL, nil)
	accountReq.Header.Set("Authorization", "Bearer "+token)

	client := http.Client{}
	accountRec, err := client.Do(accountReq)
	assert.NoError(t, err)

	// Parse the response to get the user ID
	var user struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(accountRec.Body).Decode(&user)
	assert.NoError(t, err)

	// Use the token to delete the user's own profile
	deleteURL := "http://localhost:1323/users/" + user.ID
	deleteReq, _ := http.NewRequest(http.MethodDelete, deleteURL, nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token)

	deleteRec, err := client.Do(deleteReq)
	assert.NoError(t, err)

	// Assertions for deleting user profile
	assert.Equal(t, http.StatusOK, deleteRec.StatusCode)
}

func TestUserDeleteAnother(t *testing.T) {
	token1 := signupAndLogin(t)
	token2 := signupAndLogin(t) // create another user

	// Get the second user's ID
	accountURL := "http://localhost:1323/account/me"
	accountReq, _ := http.NewRequest(http.MethodGet, accountURL, nil)
	accountReq.Header.Set("Authorization", "Bearer "+token2)

	client := http.Client{}
	accountRec, err := client.Do(accountReq)
	assert.NoError(t, err)

	// Parse the response to get the user ID
	var user struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(accountRec.Body).Decode(&user)
	assert.NoError(t, err)

	// Use the first user's token to attempt to delete the second user's profile
	deleteURL := "http://localhost:1323/users/" + user.ID
	deleteReq, _ := http.NewRequest(http.MethodDelete, deleteURL, nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token1)

	deleteRec, err := client.Do(deleteReq)
	assert.NoError(t, err)

	// Assertions for attempting to delete another user's profile
	assert.Equal(t, http.StatusForbidden, deleteRec.StatusCode)
}
