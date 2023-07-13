package handler

import (
	"fmt"
	"strings"
)

type UpdateResponse struct {
	Updated int64 `json:"updated"`
}

type DeleteResponse struct {
	Deleted int64 `json:"deleted"`
}

type ListResponse struct {
	Total int64       `json:"total"`
	Items interface{} `json:"items"`
}

func getOperatorAndValue(param string) (string, []string) {
	if strings.Contains(param, ",") {
		parts := strings.Split(param, ",")
		return parts[0], parts[1:]
	}
	return "eq", []string{param}
}

func appendQuery(query, field, op, castType string, value []string, params *[]interface{}) string {
	fieldCast := field
	if castType == "integer" {
		fieldCast = fmt.Sprintf("CAST(%s AS INTEGER)", field)
	} else if castType == "text" {
		fieldCast = fmt.Sprintf("CAST(%s AS TEXT)", field)
	}

	switch op {
	case "gt":
		query += fmt.Sprintf(" AND %s > ?", fieldCast)
		*params = append(*params, value[0])
	case "lt":
		query += fmt.Sprintf(" AND %s < ?", fieldCast)
		*params = append(*params, value[0])
	case "eq":
		query += fmt.Sprintf(" AND %s = ?", fieldCast)
		*params = append(*params, value[0])
	case "ge":
		query += fmt.Sprintf(" AND %s >= ?", fieldCast)
		*params = append(*params, value[0])
	case "le":
		query += fmt.Sprintf(" AND %s <= ?", fieldCast)
		*params = append(*params, value[0])
	case "bt":
		query += fmt.Sprintf(" AND %s BETWEEN ? AND ?", fieldCast)
		*params = append(*params, value[0], value[1])
	case "lk":
		query += fmt.Sprintf(" AND %s LIKE ?", fieldCast)
		*params = append(*params, "%"+value[0]+"%")
	}
	return query
}
