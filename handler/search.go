package handler

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Type may be one of: entry, user, city
// title:
// - entry.data.title
// - city.name
// - user.username
// slug:
// - entry.id
// - city.slug
// - user.id
type SearchResponseItem struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	Slug  string `json:"slug"`
}

// Search keyword will consider a number of fields
// - entry.data.title
// - entry.data.description
// - city.name
// - user.username
func (h *Handler) Search(c echo.Context) error {
	keyword := c.QueryParam("keyword")

	response := []SearchResponseItem{}

	// Query
	query := `
		SELECT 'entry' AS type, json_extract(data, '$.title') AS title, id AS slug FROM entries
		WHERE json_extract(data, '$.title') LIKE ? OR json_extract(data, '$.description') LIKE ?
		UNION ALL
		SELECT 'city' AS type, name AS title, slug FROM cities
		WHERE name LIKE ?
		UNION ALL
		SELECT 'user' AS type, username AS title, id AS slug FROM users
		WHERE username LIKE ?
	`

	// Params
	params := []interface{}{"%" + keyword + "%", "%" + keyword + "%", "%" + keyword + "%", "%" + keyword + "%"}

	rows, err := h.DB.Raw(query, params...).Rows()
	if err != nil {
		log.Println(err)
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var item SearchResponseItem
		err := rows.Scan(&item.Type, &item.Title, &item.Slug)
		if err != nil {
			log.Println(err)
			return err
		}
		response = append(response, item)
	}

	// TODO: Handle potential error from rows.Err()

	return c.JSON(http.StatusOK, response)
}
