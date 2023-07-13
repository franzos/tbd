package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

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
		query = appendQuery(query, "cities.country", op, "", val, &params)
	}

	if queryParams.City != "" {
		op, val := getOperatorAndValue(queryParams.City)
		query = appendQuery(query, "cities.name", op, "", val, &params)
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM entries LEFT JOIN cities ON entries.city_id = cities.id WHERE 1=1 %v", (query + ".")[:len(query)])
	query = fmt.Sprintf("SELECT entries.* FROM entries LEFT JOIN cities ON entries.city_id = cities.id WHERE 1=1 %v", query)

	query += " ORDER BY entries.created_at DESC LIMIT ? OFFSET ?"
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

		// Query for City
		cityQuery := `SELECT cities.* FROM cities WHERE cities.id = ?`
		if err := h.DB.Raw(cityQuery, entry.CityID).Scan(&entry.City).Error; err != nil {
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

	err := h.DB.Model(&model.Entry{}).Preload("CreatedBy").Preload("Files").Preload("City").First(&entry).Error
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

	// TODO: Validate data
	e := model.Entry{
		Type: s.Type,
		Data: s.Data,
	}

	// Extract city and country
	dataContent := model.BaseEntry{}
	if err := json.Unmarshal([]byte(e.Data), &dataContent); err != nil {
		log.Println(err)
	} else {
		if dataContent.Address.City != "" {
			city, err := h.GetAndCreateIfNotFoundCity(dataContent.Address)
			if err != nil {
				log.Println(err)
			} else {
				e.CityID = city.ID
			}
		}
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

	e := model.Entry{}
	if err := c.Bind(&e); err != nil {
		return err
	}

	updateData := make(map[string]interface{})
	if len(e.Data) > 0 {
		updateData["data"] = e.Data

		// TODO: Check if data is valid and has changed

		// Signature
		passphrase := []byte(os.Getenv("PGP_PASSPHRASE"))
		privateKey := user.PrivateKey

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
					updateData["data_signature"] = signed
				}
			}
		}
	}

	// Extract city state and country
	dataContent := model.BaseEntry{}
	if err := json.Unmarshal([]byte(e.Data), &dataContent); err != nil {
		log.Println(err)
	} else {
		if dataContent.Address.City != "" {
			city, err := h.GetAndCreateIfNotFoundCity(dataContent.Address)
			if err != nil {
				log.Println(err)
			} else {
				updateData["city_id"] = city.ID
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

	if len(e.Files) > 0 {
		currentEntry := model.Entry{}
		result := h.DB.Preload("Files").First(&currentEntry, "id = ?", id)
		if result.Error != nil {
			log.Println(result.Error)
			return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to fetch entry."}
		}

		h.DB.Model(&currentEntry).Association("Files").Replace(&e.Files)
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
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	if limit <= 0 {
		limit = 100
	}

	type Result struct {
		City    string `json:"city"`
		Results int    `json:"results"`
	}

	var results []Result

	if name != "" {
		h.DB.Raw(`SELECT cities.name as city, count(*) as results 
				  FROM entries 
				  INNER JOIN cities ON entries.city_id = cities.id
				  WHERE cities.name LIKE ? 
				  GROUP BY cities.name
				  LIMIT ?`, "%"+name+"%", limit).Scan(&results)
	} else {
		h.DB.Raw(`SELECT cities.name as city, count(*) as results 
				  FROM entries 
				  INNER JOIN cities ON entries.city_id = cities.id
				  GROUP BY cities.name 
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
	h.DB.Raw(`SELECT cities.country as country, count(*) as results 
			  FROM entries 
			  INNER JOIN cities ON entries.city_id = cities.id
			  GROUP BY cities.country`).Scan(&results)
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
