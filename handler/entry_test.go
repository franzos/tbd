package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"tbd/model"
	"testing"

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

func TestPostInvalidEntryType(t *testing.T) {
	token := signupAndLogin(t)

	entryData := map[string]interface{}{
		"type": "carsale",
		"data": map[string]string{
			"title": "Car Sale Entry",
		},
	}

	rec := performRequest(t, http.MethodPost, "http://localhost:1323/entries", token, entryData)
	assert.Equal(t, http.StatusBadRequest, rec.StatusCode)
}

func TestPostNewEntryAndGet(t *testing.T) {
	token := signupAndLogin(t)

	entryData := map[string]interface{}{
		"type": "apartment-short-term-rental",
		"data": map[string]interface{}{
			"title": "Some title #2",
		},
	}

	createdEntry := createEntry(t, token, entryData)

	retrievedEntry := getEntry(t, token, createdEntry.ID)

	assert.Equal(t, createdEntry.ID, retrievedEntry.ID)
	assert.Equal(t, createdEntry.Type, retrievedEntry.Type)
}

func TestPostNewEntryAndList(t *testing.T) {
	token := signupAndLogin(t)

	entryData := map[string]interface{}{
		"type": "apartment-short-term-rental",
		"data": map[string]string{
			"title": "Some title #3",
		},
	}

	createEntry(t, token, entryData)

	rec := performRequest(t, http.MethodGet, "http://localhost:1323/entries", token, nil)
	assert.Equal(t, http.StatusOK, rec.StatusCode)

	var response []struct {
		ID string `json:"id"`
	}
	err := json.NewDecoder(rec.Body).Decode(&response)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, len(response), 1)
}

func TestPostEntryWithFiles(t *testing.T) {
	token := signupAndLogin(t)

	files := uploadFiles(t, token, "../source_a4_vertical.pdf")

	entryData := map[string]interface{}{
		"type": "apartment-short-term-rental",
		"data": map[string]interface{}{
			"title": "Some title #4",
		},
		"files": files, // Use the uploaded file ID
	}

	entry := createEntry(t, token, entryData)

	retrievedEntry := getEntry(t, token, entry.ID)

	assert.Equal(t, len(files), len(retrievedEntry.Files))
}

func TestPostEntryWithFilesAndDelete(t *testing.T) {
	token := signupAndLogin(t)

	// Upload files and create an entry with those files
	files := uploadFiles(t, token, "../source_a4_vertical.pdf")
	entryData := map[string]interface{}{
		"type": "apartment-short-term-rental",
		"data": map[string]interface{}{
			"title": "Some title #4",
		},
		"files": files, // Use the uploaded file ID
	}
	entry := createEntry(t, token, entryData)

	// Retrieve the entry and verify the files are present
	retrievedEntry := getEntry(t, token, entry.ID)
	assert.Equal(t, len(files), len(retrievedEntry.Files))

	// Delete the first file
	rec := performRequest(t, http.MethodDelete, "http://localhost:1323/files/"+files[0].ID, token, nil)
	assert.Equal(t, http.StatusOK, rec.StatusCode)

	// Retrieve the entry again and verify the file is gone
	retrievedEntry = getEntry(t, token, entry.ID)
	assert.Equal(t, len(files)-1, len(retrievedEntry.Files))
}

func TestDeleteEntry(t *testing.T) {
	token := signupAndLogin(t)

	entryData := map[string]interface{}{
		"type": "apartment-short-term-rental",
		"data": map[string]string{
			"title": "Some title #5",
		},
	}

	entry := createEntry(t, token, entryData)

	rec := performRequest(t, http.MethodDelete, "http://localhost:1323/entries/"+entry.ID, token, nil)
	assert.Equal(t, http.StatusOK, rec.StatusCode)
}

func TestDeleteNonexistentEntry(t *testing.T) {
	token := signupAndLogin(t)

	rec := performRequest(t, http.MethodDelete, "http://localhost:1323/entries/nonexistent-id", token, nil)
	assert.Equal(t, http.StatusNotFound, rec.StatusCode)
}
