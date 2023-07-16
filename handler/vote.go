package handler

import (
	"fmt"
	"net/http"
	"tbd/model"

	"github.com/labstack/echo/v4"
)

func (h *Handler) FetchVotes(c echo.Context) error {
	id := c.QueryParam("id")
	tp := c.QueryParam("type")

	if tp != "entry" && tp != "comment" {
		return &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "Invalid type. Valid: entry, comment.",
		}
	}

	if id == "" {
		// TODO Check if UUID
		return &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "Invalid ID.",
		}
	}

	type Result struct {
		Up   int `json:"up"`
		Down int `json:"down"`
	}

	target := "EntryID"
	if tp == "comment" {
		target = "CommentID"
	}
	query := fmt.Sprintf(`
	SELECT 
		SUM(CASE WHEN Vote = 0 THEN 1 ELSE 0 END) as up, 
		SUM(CASE WHEN Vote = 1 THEN 1 ELSE 0 END) as down 
	FROM votes 
	WHERE %v = ?`, target)
	var result Result

	h.DB.Raw(query, id).Scan(&result)

	return c.JSON(http.StatusOK, result)
}

func (h *Handler) CastVote(c echo.Context) error {
	reqUser := c.Get("user").(*model.AuthUser)

	v := model.CastVote{}
	if err := c.Bind(&v); err != nil {
		return err
	}

	if v.Vote != 0 && v.Vote != 1 {
		return &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "Invalid vote. Valid: 0 (up), 1 (down).",
		}
	}

	if v.EntryID == "" && v.CommentID == "" {
		return &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "EntryID or CommentID is required",
		}
	}

	if v.EntryID != "" && v.CommentID != "" {
		return &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "Only one of EntryID or CommentID is allowed",
		}
	}

	voteExists := model.Vote{}
	if v.EntryID != "" {
		err := h.DB.Where("entry_id = ? AND created_by_id = ?", v.EntryID, reqUser.ID).First(&voteExists).Error
		if err == nil {
			return &echo.HTTPError{
				Code:    http.StatusBadRequest,
				Message: "Vote already casted",
			}
		}
	} else {
		err := h.DB.Where("comment_id = ? AND created_by_id = ?", v.CommentID, reqUser.ID).First(&voteExists).Error
		if err == nil {
			return &echo.HTTPError{
				Code:    http.StatusBadRequest,
				Message: "Vote already casted",
			}
		}
	}

	err := h.DB.Create(&model.Vote{
		Vote:        v.Vote,
		CreatedByID: reqUser.ID,
		EntryID:     v.EntryID,
		CommentID:   v.CommentID,
	}).Error

	if err != nil {
		return &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "Failed to cast vote",
		}
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Vote cast successfully",
	})
}

func (h *Handler) DeleteVote(c echo.Context) error {
	_, err := h.isOwnerOrAdmin(c, c.Param("id"), "vote")
	if err != nil {
		return err
	}

	id := c.Param("id")

	r := h.DB.Delete(&model.Vote{ID: id})
	if r.Error != nil {
		return &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "Failed to delete vote",
		}
	}

	if r.RowsAffected == 0 {
		return &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "Vote not found",
		}
	}

	return c.JSON(http.StatusOK, DeleteResponse{Deleted: r.RowsAffected})
}
