package handler

import (
	"fmt"
	"log"
	"net/http"
	"tbd/model"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// TODO: Improve types
func (h *Handler) isOwnerOrAdmin(c echo.Context, objectID string, objectType string) (interface{}, error) {
	reqUser := c.Get("user").(*model.AuthUser)

	var dbObject interface{}
	var errMsgs map[string]string
	if objectType == "file" {
		dbObject = &model.File{}
		errMsgs = map[string]string{
			"notFound":     "File not found.",
			"fetchFailed":  "Failed to fetch file.",
			"noPermission": "You do not have permission to update this file.",
		}
	} else if objectType == "entry" {
		dbObject = &model.Entry{}
		errMsgs = map[string]string{
			"notFound":     "Entry not found.",
			"fetchFailed":  "Failed to fetch entry.",
			"noPermission": "You do not have permission to update this entry.",
		}
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
	}

	if reqUser.IsAdmin == false && reqUser.ID != createdByID {
		log.Println(fmt.Sprintf("User is not admin and is not owner of object %v.", objectID))
		return nil, &echo.HTTPError{Code: http.StatusForbidden, Message: errMsgs["noPermission"]}
	}

	return dbObject, nil
}

func isSelfOrAdmin(c echo.Context, userID string) error {
	reqUser := c.Get("user").(*model.AuthUser)

	if reqUser.IsAdmin == false && reqUser.ID != userID {
		log.Println("User is not admin and is not owner of object.")
		// log.Println("Object ID:", dbObject.CreatedByID)
		return &echo.HTTPError{Code: http.StatusForbidden, Message: "You do not have permission to update this user."}
	}

	return nil
}
