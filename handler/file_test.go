package handler

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostNewFiles(t *testing.T) {
	token := signupAndLogin(t)

	// Create a new file upload request
	entryURL := "http://localhost:1323/files/multi"
	entryReq, _ := http.NewRequest(http.MethodPost, entryURL, nil)
	entryReq.Header.Set("Authorization", "Bearer "+token)

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
	entryReq.Header.Set("Content-Type", writer.FormDataContentType())

	// Set the request body
	entryReq.Body = ioutil.NopCloser(body)

	// Perform the request
	client := http.Client{}
	entryRec, err := client.Do(entryReq)
	assert.NoError(t, err)

	// Assertions for POST /files/multi
	assert.Equal(t, http.StatusCreated, entryRec.StatusCode)
}
