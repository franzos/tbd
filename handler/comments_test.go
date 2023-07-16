package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createComment(t *testing.T, token string, entryData map[string]interface{}) struct {
	ID string `json:"id"`
} {
	rec := performRequest(t, http.MethodPost, "http://localhost:1323/comments", token, entryData)
	assert.Equal(t, http.StatusCreated, rec.StatusCode)

	var response struct {
		ID string `json:"id"`
	}
	err := json.NewDecoder(rec.Body).Decode(&response)
	assert.NoError(t, err)

	return response
}

func updateComment(t *testing.T, token string, id string, commentData map[string]interface{}) {
	rec := performRequest(t, http.MethodPatch, "http://localhost:1323/comments/"+id, token, commentData)
	assert.Equal(t, http.StatusOK, rec.StatusCode)
}

func deleteComment(t *testing.T, token string, id string) {
	rec := performRequest(t, http.MethodDelete, "http://localhost:1323/comments/"+id, token, nil)
	assert.Equal(t, http.StatusOK, rec.StatusCode)
}

func TestCommentLifecycle(t *testing.T) {
	token := signupAndLogin(t)

	entryData := genEntryData("apartment-short-term-rental", nil)
	createdEntry := createEntry(t, token, entryData)
	retrievedEntry := getEntry(t, token, createdEntry.ID)

	// Create comment
	entryID := retrievedEntry.ID
	commentData := map[string]interface{}{
		"entry_id": entryID,
		"body":     "Test comment",
	}
	createdComment := createComment(t, token, commentData)
	commentID := createdComment.ID

	// List entry comments
	url := fmt.Sprintf("http://localhost:1323/comments?entry_id=%v", entryID)
	rec := performRequest(t, http.MethodGet, url, token, nil)
	assert.Equal(t, http.StatusOK, rec.StatusCode)

	var response = ListResponseUser{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, int(response.Total), 1)

	// Make sure commentID is found
	found := false
	comments := response.Items
	for _, comment := range comments {
		if comment.ID == commentID {
			found = true
			break
		}
	}
	assert.True(t, found, "Created comment not found in list of entry comments")

	// Update comment
	commentUpdateData := map[string]interface{}{
		"body": "Test comment updated",
	}
	updateComment(t, token, commentID, commentUpdateData)

	// TODO: Validate comment was updated

	deleteComment(t, token, commentID)

	// TODO: Validate comment was deleted
}
