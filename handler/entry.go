package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"tbd/model"
	"tbd/pgp"
)

func (h *Handler) FetchEntries(c echo.Context) error {
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	// Defaults
	if offset == 0 {
		offset = 1
	}
	if limit == 0 {
		limit = 100
	}

	count := int64(0)
	entries := []model.Entry{}
	r := h.DB.Model(&model.Entry{}).
		Preload("CreatedBy").Preload("Files").
		Order("created_at desc").
		Count(&count).
		Offset((offset - 1) * limit).
		Limit(limit).
		Find(&entries)
	if r.Error != nil {
		log.Println(r.Error)
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to fetch entries."}
	}

	return c.JSON(
		http.StatusOK,
		ListResponse{Total: int64(count), Items: responseArrFormatter[model.Entry](entries, nil, os.Getenv("DOMAIN"))},
	)
}

func (h *Handler) FetchEntry(c echo.Context) error {
	id := c.Param("id")

	var entry = model.Entry{ID: id}

	err := h.DB.Model(&model.Entry{}).Preload("CreatedBy").Preload("Files").First(&entry).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "Entry not found."}
		}
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to fetch entry."}
	}

	return c.JSON(
		http.StatusOK,
		responseFormatter[model.Entry](entry, nil, os.Getenv("DOMAIN")),
	)
}

func (h *Handler) CreateEntry(c echo.Context) error {
	reqUser := c.Get("user").(*model.AuthUser)
	user := model.User{ID: reqUser.ID}
	err := h.DB.Model(&model.User{}).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "User not found."}
		}
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to fetch user."}
	}

	s := model.SubmitEntry{}
	if err := c.Bind(&s); err != nil {
		return err
	}

	if err := c.Validate(&s); err != nil {
		log.Println(err)
		return err
	}

	e := model.Entry{
		Type: s.Type,
		Data: s.Data,
	}

	if s.Files != nil {
		e.Files = s.Files
	}

	if !e.TypeIsValid() {
		log.Printf(fmt.Sprintf("Type %s is not supported.", e.Type))
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Type is not supported."}
	}

	// If entry has files, loop over them, and make sure they exist in the DB
	// TODO: Do it with one query
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

	e.CreatedByID = reqUser.ID

	// Signature
	passphrase := []byte(os.Getenv("PGP_PASSPHRASE"))
	privateKey := user.PrivateKey

	// data to JSON string
	if privateKey != "" {
		data, err := json.Marshal(e.Data)
		if err != nil {
			// TODO: Notify admin
			log.Println(err)
		} else {
			signed, err := pgp.SignData(string(data), privateKey, passphrase)
			if err != nil {
				// TODO: Notify admin
				log.Println(err)
			} else {
				e.DataSignature = signed
			}
		}
	}

	r := h.DB.Create(&e)
	if r.Error != nil {
		log.Println(r.Error)
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to create entry."}
	}

	provErr := h.markFilesAsProvisioned(e.Files)
	if provErr != nil {
		log.Println(err)
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to mark files as provisioned."}
	}

	return c.JSON(http.StatusCreated, e)
}

func (h *Handler) UpdateEntry(c echo.Context) error {
	_, err := h.isOwnerOrAdmin(c, c.Param("id"), "entry")
	if err != nil {
		return err
	}
	user := model.User{}
	if err := h.DB.Model(&model.User{}).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "User not found."}
		}
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to fetch user."}
	}

	id := c.Param("id")

	entry := model.Entry{}
	if err := c.Bind(&entry); err != nil {
		return err
	}

	updateData := make(map[string]interface{})
	if len(entry.Data) > 0 {
		updateData["data"] = entry.Data

		// Signature
		passphrase := []byte(os.Getenv("PGP_PASSPHRASE"))
		privateKey := user.PrivateKey

		// data to JSON string
		if privateKey != "" {
			// JSON string from e.Data; not bytes
			// so that this works across platfoms (as in, someone can generate a string in JS, pyhton, etc. and it will work)
			data, err := json.Marshal(entry.Data)
			if err != nil {
				// TODO: Notify admin
				log.Println(err)
			} else {
				signed, err := pgp.SignData(string(data), privateKey, passphrase)
				if err != nil {
					// TODO: Notify admin
					log.Println(err)
				} else {
					updateData["data_signature"] = signed
				}
			}
		}
	}

	if len(updateData) > 0 {
		r := h.DB.Model(&model.Entry{ID: id}).Updates(updateData)
		if r.Error != nil {
			log.Println(r.Error)
			return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to update entry."}
		}
	}

	if len(entry.Files) > 0 {
		currentEntry := model.Entry{}
		result := h.DB.Preload("Files").First(&currentEntry, "id = ?", id)
		if result.Error != nil {
			log.Println(result.Error)
			return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to fetch entry."}
		}

		h.DB.Model(&currentEntry).Association("Files").Replace(&entry.Files)
	}

	return c.JSON(http.StatusOK, UpdateResponse{Updated: 1}) // Assume 1 row affected since the entry exists and you're here.
}

func (h *Handler) DeleteEntry(c echo.Context) error {
	_, err := h.isOwnerOrAdmin(c, c.Param("id"), "entry")
	if err != nil {
		return err
	}

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

func (h *Handler) CountCityResults(c echo.Context) error {

	type Result struct {
		City    string
		Results int
	}

	var results []Result
	h.DB.Raw(`SELECT json_extract(data, '$.location.city') as city, count(*) as results FROM entries GROUP BY city`).Scan(&results)
	return c.JSON(http.StatusOK, results)
}
