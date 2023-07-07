package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
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

func uploadFiles(t *testing.T, token, filepath string) []model.File {
	// Create a new file upload request
	uploadURL := "http://localhost:1323/files/multi"
	uploadReq, _ := http.NewRequest(http.MethodPost, uploadURL, nil)
	uploadReq.Header.Set("Authorization", "Bearer "+token)

	// Create a new form data buffer
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Open the file
	file, err := os.Open(filepath)
	assert.NoError(t, err)
	defer file.Close()

	// Create a form file field
	part, err := writer.CreateFormFile("files", file.Name())
	assert.NoError(t, err)

	// Copy the file data to the form file field
	_, err = io.Copy(part, file)
	assert.NoError(t, err)

	// Close the writer
	err = writer.Close()
	assert.NoError(t, err)

	// Set the content type header
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())

	// Set the request body
	uploadReq.Body = ioutil.NopCloser(body)

	// Perform the request to upload the file
	client := http.Client{}
	uploadRec, err := client.Do(uploadReq)
	assert.NoError(t, err)

	// Assertions for POST /files/multi
	assert.Equal(t, http.StatusCreated, uploadRec.StatusCode)

	// Parse the upload response which is a JSON array of files
	var uploadResponse struct {
		Files []model.File `json:"files"`
	}

	err = json.NewDecoder(uploadRec.Body).Decode(&uploadResponse)
	assert.NoError(t, err)

	return uploadResponse.Files
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
