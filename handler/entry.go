package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"tbd/model"
)

var entryTypes = []string{
	"apartment-short-term-rental",
	"apartment-long-term-rental",
	"apartment-sale",
}

func validateEntryType(t string) bool {
	for _, v := range entryTypes {
		if v == t {
			return true
		}
	}
	return false
}

func (h *Handler) CreateEntry(c echo.Context) (err error) {
	u, err := userFromToken(c)
	if err != nil {
		log.Printf("error: %v", err)
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to parse provided token."}
	}

	s := model.SubmitEntry{}
	if err = c.Bind(&s); err != nil {
		return
	}

	if err = c.Validate(&s); err != nil {
		fmt.Println(err)
		return err
	}

	e := model.Entry{
		Type: s.Type,
		Data: s.Data,
	}

	if !validateEntryType(e.Type) {
		log.Printf("Type is not supported.")
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Type is not supported."}
	}

	// If entry has files, loop over them, and make sure they exist in the DB
	if len(e.Files) > 0 {
		for i, f := range e.Files {
			if f.ID == "" {
				log.Printf(fmt.Sprintf("File %d is incomplete.", i+1))
				return &echo.HTTPError{Code: http.StatusBadRequest, Message: fmt.Sprintf("File %d is incomplete.", i+1)}
			}

			r := h.DB.First(&f)
			if r.Error != nil {
				log.Printf(fmt.Sprintf("File %v does not exist.", f.ID))
				return &echo.HTTPError{Code: http.StatusBadRequest, Message: fmt.Sprintf("File %v does not exist.", f.ID)}
			}
		}
	}

	e.CreatedBy = u

	r := h.DB.Create(&e)
	if r.Error != nil {
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to create entry."}
	}

	return c.JSON(http.StatusCreated, e)
}

func (h *Handler) UpdateEntry(c echo.Context) (err error) {
	if err != nil {
		log.Printf("error: %v", err)
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to parse provided token."}
	}

	id := c.Param("id")
	entry := model.Entry{}
	if err = c.Bind(&entry); err != nil {
		return
	}

	// We really only allow updating the data field for now
	r := h.DB.Model(model.Entry{ID: id}).Update("data", entry.Data)
	if r.Error != nil {
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to update entry."}
	}

	return c.JSON(http.StatusOK, UpdateResponse{Updated: r.RowsAffected})
}

func (h *Handler) FetchEntries(c echo.Context) (err error) {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	// Defaults
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 100
	}

	entries := []model.Entry{}
	error := h.DB.Model(&model.Entry{}).Preload("CreatedBy").Order("created_at desc").Offset((page - 1) * limit).Limit(limit).Find(&entries).Error
	if error != nil {
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to fetch entries."}
	}

	return c.JSON(http.StatusOK, entries)
}

func (h *Handler) FetchEntry(c echo.Context) error {
	id := c.Param("id")

	log.Println(id)

	var entry = model.Entry{ID: id}
	err := h.DB.First(&entry).Error
	if err != nil {
		return &echo.HTTPError{Code: http.StatusNotFound, Message: "Entry not found."}
	}

	log.Println(entry.ID)

	return c.JSON(http.StatusOK, entry)
}

func (h *Handler) DeleteEntry(c echo.Context) (err error) {
	id := c.Param("id")

	r := h.DB.Delete(model.Entry{ID: id})
	if r.Error != nil {
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to delete entry."}
	}

	if r.RowsAffected == 0 {
		return &echo.HTTPError{Code: http.StatusNotFound, Message: "Entry not found."}
	}

	return c.JSON(http.StatusOK, DeleteResponse{Deleted: r.RowsAffected})
}
