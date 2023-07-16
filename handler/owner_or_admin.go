package handler

import (
	"fmt"
	"log"
	"net/http"
	"tbd/model"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func errorMsgs(name string) map[string]string {
	msgs := map[string]string{
		"notFound":     fmt.Sprintf("%v not found.", name),
		"fetchFailed":  fmt.Sprintf("Failed to fetch %v.", name),
		"noPermission": fmt.Sprintf("You do not have permission to change this %v.", name),
	}
	return msgs
}

// Runs 3 checks
// 1. Check if the UUID is valid
// 2. Check if the object exists
// 3. Check if the user is the owner of the object or an admin

// TODO: Improve types
func (h *Handler) isOwnerOrAdmin(c echo.Context, objectID string, objectType string) (interface{}, error) {
	reqUser := c.Get("user").(*model.AuthUser)

	if _, err := uuid.Parse(objectID); err != nil {
		return nil, &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid UUID."}
	}

	var dbObject interface{}
	var errMsgs map[string]string
	if objectType == "file" {
		dbObject = &model.File{}
		errMsgs = errorMsgs("file")
	} else if objectType == "entry" {
		dbObject = &model.Entry{}
		errMsgs = errorMsgs("entry")
	} else if objectType == "vote" {
		dbObject = &model.Vote{}
		errMsgs = errorMsgs("vote")
	} else if objectType == "comment" {
		dbObject = &model.Comment{}
		errMsgs = errorMsgs("comment")
	} else {
		return nil, &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid object type."}
	}

	err := h.DB.Model(dbObject).Preload("CreatedBy").First(dbObject, "id = ?", objectID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &echo.HTTPError{Code: http.StatusNotFound, Message: errMsgs["notFound"]}
		}
		log.Println(err)
		return nil, &echo.HTTPError{Code: http.StatusInternalServerError, Message: errMsgs["fetchFailed"]}
	}

	createdByID := ""
	switch v := dbObject.(type) {
	case *model.File:
		createdByID = v.CreatedByID
	case *model.Entry:
		createdByID = v.CreatedByID
	case *model.Vote:
		createdByID = v.CreatedByID
	case *model.Comment:
		createdByID = v.CreatedByID
	}

	if reqUser.IsAdmin == false && reqUser.ID != createdByID {
		log.Println(fmt.Sprintf("User is not admin and is not owner of object %v.", objectID))
		return nil, &echo.HTTPError{Code: http.StatusForbidden, Message: errMsgs["noPermission"]}
	}

	return dbObject, nil
}

func isSelfOrAdmin(c echo.Context, userID string) error {
	reqUser := c.Get("user").(*model.AuthUser)

	if _, err := uuid.Parse(userID); err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid UUID."}
	}

	if reqUser.IsAdmin == false && reqUser.ID != userID {
		log.Println("User is not admin and is not owner of object.")
		return &echo.HTTPError{Code: http.StatusForbidden, Message: "You do not have permission to update this user."}
	}

	return nil
}
