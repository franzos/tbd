package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"tbd/model"
	"testing"

	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
)

func performRequest(t *testing.T, method, url, token string, data interface{}) *http.Response {
	payload, _ := json.Marshal(data)

	req, _ := http.NewRequest(method, url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := http.Client{}
	rec, err := client.Do(req)
	assert.NoError(t, err)

	return rec
}

func getEntry(t *testing.T, token string, id string) model.Entry {
	getURL := "http://localhost:1323/entries/" + id
	getReq, _ := http.NewRequest(http.MethodGet, getURL, nil)
	getReq.Header.Set("Authorization", "Bearer "+token)

	client := http.Client{}
	getRec, err := client.Do(getReq)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, getRec.StatusCode)

	// Validate the response body
	var getResponse model.Entry
	err = json.NewDecoder(getRec.Body).Decode(&getResponse)
	assert.NoError(t, err)

	getRec.Body.Close()

	return getResponse
}

func createEntry(t *testing.T, token string, entryData map[string]interface{}) struct {
	ID   string `json:"id"`
	Type string `json:"type"`
} {
	rec := performRequest(t, http.MethodPost, "http://localhost:1323/entries", token, entryData)
	assert.Equal(t, http.StatusCreated, rec.StatusCode)

	var response struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}
	err := json.NewDecoder(rec.Body).Decode(&response)
	assert.NoError(t, err)

	return response
}

func updateEntry(t *testing.T, token string, entryID string, entryData map[string]interface{}) {
	entryURL := fmt.Sprintf("http://localhost:1323/entries/%s", entryID)

	rec := performRequest(t, http.MethodPatch, entryURL, token, entryData)

	if rec.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK, got: %v", rec.StatusCode)
	}
}

func genEntryData(listingType string, files []model.File) map[string]interface{} {
	fake := faker.New()
	entryData := map[string]interface{}{
		"type": listingType,
		"data": map[string]interface{}{
			"title":       fake.Lorem().Text(60),
			"description": fake.Lorem().Sentence(8),
		},
	}
	if files != nil {
		entryData["files"] = files
	}
	return entryData
}

func TestPostInvalidEntryType(t *testing.T) {
	token := signupAndLogin(t)

	entryData := genEntryData("carsale", nil)

	rec := performRequest(t, http.MethodPost, "http://localhost:1323/entries", token, entryData)
	assert.Equal(t, http.StatusBadRequest, rec.StatusCode)
}

func TestPostNewEntryWithoutFilesAndGet(t *testing.T) {
	token := signupAndLogin(t)

	entryData := genEntryData("apartment-short-term-rental", nil)
	createdEntry := createEntry(t, token, entryData)
	retrievedEntry := getEntry(t, token, createdEntry.ID)

	assert.Equal(t, createdEntry.ID, retrievedEntry.ID)
	assert.Equal(t, createdEntry.Type, retrievedEntry.Type)
}

func TestPostNewEntryWithoutFilesAndList(t *testing.T) {
	token := signupAndLogin(t)

	entryData := genEntryData("apartment-short-term-rental", nil)
	createEntry(t, token, entryData)

	rec := performRequest(t, http.MethodGet, "http://localhost:1323/entries", token, nil)
	assert.Equal(t, http.StatusOK, rec.StatusCode)

	var response = ListResponse{
		Items: []model.PublicEntry{},
		Total: 0,
	}
	err := json.NewDecoder(rec.Body).Decode(&response)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, int(response.Total), 1)
}

func TestPostEntryWithFiles(t *testing.T) {
	token := signupAndLogin(t)

	files := uploadFiles(t, token, "../concorde.jpg")
	entryData := genEntryData("apartment-short-term-rental", files)
	createdEntry := createEntry(t, token, entryData)
	retrievedEntry := getEntry(t, token, createdEntry.ID)

	assert.Equal(t, len(files), len(retrievedEntry.Files))
}

func TestPostEntryWithFilesAndUpdate(t *testing.T) {
	token := signupAndLogin(t)

	files := uploadFiles(t, token, "../concorde.jpg")
	fake := faker.New()
	entryData := genEntryData("apartment-short-term-rental", files)
	createdEntry := createEntry(t, token, entryData)
	retrievedEntry := getEntry(t, token, createdEntry.ID)

	assert.Equal(t, len(files), len(retrievedEntry.Files))

	// Update the entry title
	newTitle := fake.Lorem().Text(40)
	updateData := map[string]interface{}{
		"data": map[string]interface{}{
			"title":       newTitle,
			"description": entryData["data"].(map[string]interface{})["description"],
		},
	}

	updateEntry(t, token, createdEntry.ID, updateData)
	updatedEntry := getEntry(t, token, createdEntry.ID)

	dataMap := make(map[string]interface{})
	err := json.Unmarshal(updatedEntry.Data, &dataMap)
	if err != nil {
		t.Fatalf("Error unmarshaling JSON: %v", err)
	}

	title, ok := dataMap["title"].(string)
	if !ok {
		t.Fatalf("Title not found or not a string")
	}

	// Now you can use the `title` in your assert
	assert.Equal(t, newTitle, title)
}

func TestPostEntryWithFilesAndUpdateUnauthorizedUser(t *testing.T) {
	token := signupAndLogin(t)

	files := uploadFiles(t, token, "../concorde.jpg")
	entryData := genEntryData("apartment-short-term-rental", files)
	createdEntry := createEntry(t, token, entryData)
	retrievedEntry := getEntry(t, token, createdEntry.ID)

	assert.Equal(t, len(files), len(retrievedEntry.Files))

	// Add another user
	anotherUserToken := signupAndLogin(t)

	// Try to update the entry with another user
	updateData := map[string]interface{}{
		"data": map[string]interface{}{
			"title": "This should not work",
		},
	}

	resp := performRequest(t, http.MethodPatch, "http://localhost:1323/entries/"+createdEntry.ID, anotherUserToken, updateData)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestPostEntryWithFilesAndDelete(t *testing.T) {
	token := signupAndLogin(t)

	// Upload files and create an entry with those files
	files := uploadFiles(t, token, "../concorde.jpg")

	entryData := genEntryData("apartment-short-term-rental", files)
	createdEntry := createEntry(t, token, entryData)
	retrievedEntry := getEntry(t, token, createdEntry.ID)
	assert.Equal(t, len(files), len(retrievedEntry.Files))

	// Delete the first file
	rec := performRequest(t, http.MethodDelete, "http://localhost:1323/files/"+files[0].ID, token, nil)
	assert.Equal(t, http.StatusOK, rec.StatusCode)

	// Retrieve the entry again and verify the file is gone
	retrievedEntry = getEntry(t, token, createdEntry.ID)
	assert.Equal(t, len(files)-1, len(retrievedEntry.Files))
}

func TestDeleteEntry(t *testing.T) {
	token := signupAndLogin(t)

	entryData := genEntryData("apartment-short-term-rental", nil)
	createdEntry := createEntry(t, token, entryData)

	rec := performRequest(t, http.MethodDelete, "http://localhost:1323/entries/"+createdEntry.ID, token, nil)
	assert.Equal(t, http.StatusOK, rec.StatusCode)
}

func TestDeleteEntryWithUnauthorizedUser(t *testing.T) {
	token := signupAndLogin(t)

	entryData := genEntryData("apartment-short-term-rental", nil)
	createdEntry := createEntry(t, token, entryData)
	newUserToken := signupAndLogin(t)

	rec := performRequest(t, http.MethodDelete, "http://localhost:1323/entries/"+createdEntry.ID, newUserToken, nil)
	assert.Equal(t, http.StatusForbidden, rec.StatusCode)
}

func TestDeleteNonexistentEntry(t *testing.T) {
	token := signupAndLogin(t)

	rec := performRequest(t, http.MethodDelete, "http://localhost:1323/entries/nonexistent-id", token, nil)
	assert.Equal(t, http.StatusNotFound, rec.StatusCode)
}
