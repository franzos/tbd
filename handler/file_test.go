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

func TestPostNewFiles(t *testing.T) {
	token := signupAndLogin(t)

	files := uploadFiles(t, token, "../source_a4_vertical.pdf")
	assert.GreaterOrEqual(t, len(files), 1)
}

func TestFileLifecycle(t *testing.T) {
	token := signupAndLogin(t)

	// Upload new files
	files := uploadFiles(t, token, "../source_a4_vertical.pdf")
	assert.GreaterOrEqual(t, len(files), 1)

	// For the sake of this example, we will operate on the first file
	file := files[0]

	// Delete the file
	rec := performRequest(t, http.MethodDelete, "http://localhost:1323/files/"+file.ID, token, nil)
	assert.Equal(t, http.StatusOK, rec.StatusCode)
}
