package handler

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
