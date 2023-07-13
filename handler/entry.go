package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"tbd/model"
	"tbd/pgp"
)

func (h *Handler) FetchEntries(c echo.Context) error {
	queryParams := new(model.EntryQueryParams)
	if err := (&echo.DefaultBinder{}).BindQueryParams(c, queryParams); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Validate query params
	var validate = validator.New()
	if err := validate.Struct(queryParams); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// defaults
	if queryParams.Offset < 0 {
		queryParams.Offset = 0
	}
	if queryParams.Limit < 1 {
		queryParams.Limit = 20
	}

	var entries []model.Entry
	var count int64
	var params []interface{}

	query := ""

	if queryParams.Type != "" {
		query += " AND type = ?"
		params = append(params, queryParams.Type)
	}
	fmt.Println("Parameters TYPE: ", params)

	if queryParams.Price != "" {
		op, val := getOperatorAndValue(queryParams.Price)
		query = appendQuery(query, "JSON_EXTRACT(data, '$.price')", op, "integer", val, &params)
	}
	fmt.Println("Parameters PRICE: ", params)

	if queryParams.StartDate != "" {
		op, val := getOperatorAndValue(queryParams.StartDate)
		query = appendQuery(query, "JSON_EXTRACT(data, '$.start_date')", op, "", val, &params)
	}

	if queryParams.EndDate != "" {
		op, val := getOperatorAndValue(queryParams.EndDate)
		query = appendQuery(query, "JSON_EXTRACT(data, '$.end_date')", op, "integer", val, &params)
	}

	if queryParams.Country != "" {
		op, val := getOperatorAndValue(queryParams.Country)
		query = appendQuery(query, "JSON_EXTRACT(data, '$.address.country')", op, "", val, &params)
	}

	if queryParams.City != "" {
		op, val := getOperatorAndValue(queryParams.City)
		query = appendQuery(query, "JSON_EXTRACT(data, '$.address.city')", op, "", val, &params)
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM entries WHERE 1=1 %v", (query + ".")[:len(query)])
	query = fmt.Sprintf("SELECT * FROM entries WHERE 1=1 %v", query)

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	params = append(params, queryParams.Limit, queryParams.Offset)

	fmt.Println("Final query: ", query)
	fmt.Println("Parameters: ", params)

	// Run the queries
	rows, err := h.DB.Raw(query, params...).Rows()
	if err != nil {
		log.Println(err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var entry model.Entry

		if err := h.DB.ScanRows(rows, &entry); err != nil {
			log.Println(err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		// Query for Files
		fileQuery := `SELECT files.* FROM files
        INNER JOIN entry_files ON entry_files.file_id = files.id
        WHERE entry_files.entry_id = ?`
		if err := h.DB.Raw(fileQuery, entry.ID).Find(&entry.Files).Error; err != nil {
			log.Println(err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		// Query for CreatedBy
		createdByQuery := `SELECT users.* FROM users WHERE users.id = ?`
		if err := h.DB.Raw(createdByQuery, entry.CreatedByID).Scan(&entry.CreatedBy).Error; err != nil {
			log.Println(err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		entries = append(entries, entry)
	}

	if err := h.DB.Raw(countQuery, params...).Count(&count).Error; err != nil {
		log.Println(err)
		return err
	}

	return c.JSON(
		http.StatusOK,
		ListResponse{Total: count, Items: responseArrFormatter[model.Entry](entries, nil, os.Getenv("DOMAIN"))},
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

func (h *Handler) EntriesByCity(c echo.Context) error {
	// Filter result by name if provided (LIKE)
	name := c.QueryParam("name")
	limit := c.QueryParam("limit")

	if limit == "" {
		limit = "100"
	}

	type Result struct {
		City    string `json:"city"`
		Results int    `json:"results"`
	}

	var results []Result

	if name != "" {
		h.DB.Raw(`SELECT json_extract(data, '$.location.city') as city, count(*) as results 
				FROM entries 
				WHERE json_extract(data, '$.location.city') LIKE ? 
				GROUP BY city 
				LIMIT ?`, "%"+name+"%", limit).Scan(&results)
	} else {
		h.DB.Raw(`SELECT json_extract(data, '$.location.city') as city, count(*) as results 
				FROM entries 
				GROUP BY city 
				LIMIT ?`, limit).Scan(&results)
	}
	return c.JSON(http.StatusOK, results)
}

func (h *Handler) EntriesByCountry(c echo.Context) error {

	type Result struct {
		Country string `json:"country"`
		Results int    `json:"results"`
	}

	var results []Result
	h.DB.Raw(`SELECT json_extract(data, '$.location.country') as country, count(*) as results FROM entries GROUP BY country`).Scan(&results)
	return c.JSON(http.StatusOK, results)
}

func (h *Handler) EntriesByType(c echo.Context) error {
	type Result struct {
		Type    string `json:"type"`
		Results int    `json:"results"`
	}

	// Extract query parameters
	city := c.QueryParam("city")
	country := c.QueryParam("country")

	var results []Result
	var query string
	var params []interface{}

	// Construct the SQL query based on the query parameters
	if city != "" {
		query = `SELECT type, COUNT(*) AS results FROM entries WHERE JSON_EXTRACT(data, '$.address.city') = ? GROUP BY type`
		params = append(params, city)
	} else if country != "" {
		query = `SELECT type, COUNT(*) AS results FROM entries WHERE JSON_EXTRACT(data, '$.address.country') = ? GROUP BY type`
		params = append(params, country)
	} else {
		query = `SELECT type, COUNT(*) AS results FROM entries GROUP BY type`
	}

	h.DB.Raw(query, params...).Scan(&results)
	return c.JSON(http.StatusOK, results)
}
