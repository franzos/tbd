package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"tbd/model"
	"testing"
	"time"

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

func randate() time.Time {
	min := time.Date(2020, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2030, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	delta := max - min

	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0)
}

func genEntryData(listingType string, files []model.File) map[string]interface{} {
	fake := faker.New()

	fakePrice := fmt.Sprintf("%.2f", float64(fake.Currency().Number()))

	fakeAddress := fake.Address()
	address := model.Address{
		Street:   fakeAddress.StreetAddress(),
		City:     fakeAddress.City(),
		PostCode: fakeAddress.PostCode(),
		State:    fakeAddress.State(),
		Country:  fakeAddress.Country(),
	}

	baseEntry := model.BaseEntry{
		Title:       fake.Lorem().Text(60),
		Description: fake.Lorem().Sentence(8),
		Address:     address,
	}

	var entry interface{}

	switch listingType {
	case "apartment-short-term-rental":
		entry = model.EntryApartmentShortTermRental{
			BaseEntry:     baseEntry,
			StartDate:     randate().Format(time.RFC3339),
			EndDate:       randate().Format(time.RFC3339),
			Price:         fakePrice,
			PriceInterval: "month",
		}
	case "apartment-long-term-rental":
		entry = model.EntryApartmentLongTermRental{
			BaseEntry:     baseEntry,
			StartDate:     randate().Format(time.RFC3339),
			Price:         fakePrice,
			PriceInterval: "month",
		}
	case "pet-sitter":
		entry = model.EntryPetSitter{
			BaseEntry:     baseEntry,
			Price:         fakePrice,
			PriceInterval: "hour",
		}
	case "item-sale":
		entry = model.EntryItemSale{
			BaseEntry: baseEntry,
			Price:     fakePrice,
		}
	case "looking-for":
		entry = model.EntryLookingFor{
			BaseEntry: baseEntry,
		}
	default:
		panic("invalid listing type")
	}

	entryData := map[string]interface{}{
		"type": listingType,
		"data": entry,
	}

	if files != nil {
		entryData["files"] = files
	}

	return entryData
}

func TestPostInvalidEntryType(t *testing.T) {
	token := signupAndLogin(t)

	rec := performRequest(t, http.MethodPost, "http://localhost:1323/entries", token, model.Entry{
		Type: "carsale",
		Data: nil,
	})
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

func TestPostEntryWithFilesPetSitter(t *testing.T) {
	token := signupAndLogin(t)

	files := uploadFiles(t, token, "../concorde.jpg")
	entryData := genEntryData("pet-sitter", files)
	createdEntry := createEntry(t, token, entryData)
	retrievedEntry := getEntry(t, token, createdEntry.ID)

	assert.Equal(t, len(files), len(retrievedEntry.Files))
}

func TestPostEntryWithFilesItemSale(t *testing.T) {
	token := signupAndLogin(t)

	files := uploadFiles(t, token, "../concorde.jpg")
	entryData := genEntryData("item-sale", files)
	createdEntry := createEntry(t, token, entryData)
	retrievedEntry := getEntry(t, token, createdEntry.ID)

	assert.Equal(t, len(files), len(retrievedEntry.Files))
}

func TestPostEntryWithFilesLookingFor(t *testing.T) {
	token := signupAndLogin(t)

	files := uploadFiles(t, token, "../concorde.jpg")
	entryData := genEntryData("looking-for", files)
	createdEntry := createEntry(t, token, entryData)
	retrievedEntry := getEntry(t, token, createdEntry.ID)

	assert.Equal(t, len(files), len(retrievedEntry.Files))
}

func TestPostEntryWithFilesAndUpdate(t *testing.T) {
	token := signupAndLogin(t)

	files := uploadFiles(t, token, "../concorde.jpg")
	fake := faker.New()
	entryData := genEntryData("apartment-short-term-rental", files)

	// Cast the return value of genEntryData to the appropriate type
	aptShortTermRental, ok := entryData["data"].(model.EntryApartmentShortTermRental)
	if !ok {
		t.Fatalf("Invalid entry data type")
	}

	// Create the entry
	createdEntry := createEntry(t, token, entryData)
	retrievedEntry := getEntry(t, token, createdEntry.ID)

	assert.Equal(t, len(files), len(retrievedEntry.Files))

	// Update the entry title
	newTitle := fake.Lorem().Text(40)
	updateData := map[string]interface{}{
		"data": map[string]interface{}{
			"title":       newTitle,
			"description": aptShortTermRental.Description,
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
