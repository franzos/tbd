package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"tbd/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func postEntry(t *testing.T, token string, entryData map[string]interface{}) *http.Response {
	entryURL := "http://localhost:1323/entries"
	entryPayload, _ := json.Marshal(entryData)

	entryReq, _ := http.NewRequest(http.MethodPost, entryURL, bytes.NewBuffer(entryPayload))
	entryReq.Header.Set("Content-Type", "application/json")
	entryReq.Header.Set("Authorization", "Bearer "+token)

	client := http.Client{}
	entryRec, err := client.Do(entryReq)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusCreated, entryRec.StatusCode)

	return entryRec
}

func TestPostNewEntry(t *testing.T) {
	token := signupAndLogin(t)

	entryData := map[string]interface{}{
		"type": "apartment-short-term-rental",
		"data": map[string]string{
			"title": "Some title",
		},
	}

	postEntry(t, token, entryData)
}

func TestPostInvalidEntryType(t *testing.T) {
	token := signupAndLogin(t)

	entryData := map[string]interface{}{
		"type": "carsale",
		"data": map[string]string{
			"title": "Car Sale Entry",
		},
	}

	entryURL := "http://localhost:1323/entries"
	entryPayload, _ := json.Marshal(entryData)

	entryReq, _ := http.NewRequest(http.MethodPost, entryURL, bytes.NewBuffer(entryPayload))
	entryReq.Header.Set("Content-Type", "application/json")
	entryReq.Header.Set("Authorization", "Bearer "+token)

	client := http.Client{}
	entryRec, err := client.Do(entryReq)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, entryRec.StatusCode)
}

func TestPostNewEntryAndGet(t *testing.T) {
	token := signupAndLogin(t)

	entryData := map[string]interface{}{
		"type": "apartment-short-term-rental",
		"data": map[string]string{
			"title": "Some title",
		},
	}

	// Parse the upload response which is a JSON array of files
	r := postEntry(t, token, entryData)
	response := model.Entry{}
	err := json.NewDecoder(r.Body).Decode(&response)
	assert.NoError(t, err)

	getURL := "http://localhost:1323/entries/" + response.ID
	getReq, _ := http.NewRequest(http.MethodGet, getURL, nil)

	client := http.Client{}
	getRec, err := client.Do(getReq)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, getRec.StatusCode)

	// Validate the response body
	var getResponse struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(getRec.Body).Decode(&getResponse)
	assert.NoError(t, err)

	// Perform additional assertions on the response body
	assert.Equal(t, response.ID, getResponse.ID)

	// Close the response body
	getRec.Body.Close()
}

func TestPostNewEntryAndList(t *testing.T) {
	token := signupAndLogin(t)

	entryData := map[string]interface{}{
		"type": "apartment-short-term-rental",
		"data": map[string]string{
			"title": "Some title",
		},
	}

	postEntry(t, token, entryData)

	listURL := "http://localhost:1323/entries"
	listReq, _ := http.NewRequest(http.MethodGet, listURL, nil)

	client := http.Client{}
	listRec, err := client.Do(listReq)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, listRec.StatusCode)

	// Validate the response body
	var listResponse []struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(listRec.Body).Decode(&listResponse)
	assert.NoError(t, err)

	// Perform additional assertions on the response body
	assert.GreaterOrEqual(t, len(listResponse), 1)

	// Close the response body
	listRec.Body.Close()
}

func TestPostEntryWithFiles(t *testing.T) {
	token := signupAndLogin(t)

	// Create a new file upload request
	uploadURL := "http://localhost:1323/files/multi"
	uploadReq, _ := http.NewRequest(http.MethodPost, uploadURL, nil)
	uploadReq.Header.Set("Authorization", "Bearer "+token)

	// Create a new form data buffer
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Open the file
	file, err := os.Open("../source_a4_vertical.pdf")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	// Create a form file field
	part, err := writer.CreateFormFile("files", filepath.Base(file.Name()))
	if err != nil {
		t.Fatal(err)
	}

	// Copy the file data to the form file field
	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatal(err)
	}

	// Close the writer
	err = writer.Close()
	if err != nil {
		t.Fatal(err)
	}

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

	// Create the entry using the uploaded file
	entryURL := "http://localhost:1323/entries"
	entryData := map[string]interface{}{
		"type": "apartment-short-term-rental",
		"data": map[string]interface{}{
			"title": "Some title",
		},
		"files": uploadResponse.Files, // Use the uploaded file ID
	}
	entryPayload, _ := json.Marshal(entryData)

	entryReq, _ := http.NewRequest(http.MethodPost, entryURL, bytes.NewBuffer(entryPayload))
	entryReq.Header.Set("Content-Type", "application/json")
	entryReq.Header.Set("Authorization", "Bearer "+token)

	// Perform the request to create the entry
	entryRec, err := client.Do(entryReq)
	assert.NoError(t, err)

	// Assertions for POST /entries
	assert.Equal(t, http.StatusCreated, entryRec.StatusCode)
}

func TestDeleteEntry(t *testing.T) {
	token := signupAndLogin(t)

	// Create a new entry
	entryURL := "http://localhost:1323/entries"
	entryData := map[string]interface{}{
		"type": "apartment-short-term-rental",
		"data": map[string]string{
			"title": "Some title",
		},
	}
	entryPayload, _ := json.Marshal(entryData)

	entryReq, _ := http.NewRequest(http.MethodPost, entryURL, bytes.NewBuffer(entryPayload))
	entryReq.Header.Set("Content-Type", "application/json")
	entryReq.Header.Set("Authorization", "Bearer "+token)

	client := http.Client{}
	entryRec, err := client.Do(entryReq)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusCreated, entryRec.StatusCode)

	// Parse the entry response
	var entryResponse struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(entryRec.Body).Decode(&entryResponse)
	assert.NoError(t, err)

	// Delete the entry
	deleteURL := fmt.Sprintf("http://localhost:1323/entries/%s", entryResponse.ID)
	deleteReq, _ := http.NewRequest(http.MethodDelete, deleteURL, nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token)

	deleteRec, err := client.Do(deleteReq)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, deleteRec.StatusCode)
}

func TestDeleteNonexistentEntry(t *testing.T) {
	token := signupAndLogin(t)

	// Delete a nonexistent entry
	deleteURL := "http://localhost:1323/entries/nonexistent-id"
	deleteReq, _ := http.NewRequest(http.MethodDelete, deleteURL, nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token)

	client := http.Client{}
	deleteRec, err := client.Do(deleteReq)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, deleteRec.StatusCode)
}
