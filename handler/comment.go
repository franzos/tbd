package handler

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"tbd/model"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func (h *Handler) FetchComments(c echo.Context) error {
	entryID := c.QueryParam("entry_id")

	if _, err := uuid.Parse(entryID); err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "You need to supply an entry ID (entry_id) query param to fetch comments."}
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit == 0 {
		limit = 20
	}

	comments := []model.Comment{}
	var count int64
	err := h.DB.Model(&model.Comment{}).
		Preload("CreatedBy").
		Where("entry_id = ?", entryID).
		Count(&count).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&comments).Error
	if err != nil {
		log.Println(err)
		return &echo.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to fetch comments.",
		}
	}

	return c.JSON(
		http.StatusOK,
		ListResponse{
			Total: count,
			Items: responseArrFormatter[model.Comment](comments, nil, os.Getenv("DOMAIN")),
		},
	)
}

func (h *Handler) MakeComment(c echo.Context) error {
	reqUser := c.Get("user").(*model.AuthUser)
	user := model.User{ID: reqUser.ID}
	err := h.DB.Model(&model.User{}).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "User not found."}
		}
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to fetch user."}
	}

	v := model.MakeComment{}
	if err := c.Bind(&v); err != nil {
		return err
	}

	if err := c.Validate(&v); err != nil {
		log.Println(err)
		return err
	}

	if v.EntryID == "" {
		return &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "Invalid ID.",
		}
	}

	// TODO: Sign

	comment := model.Comment{
		EntryID:     v.EntryID,
		Body:        v.Body,
		CreatedByID: user.ID,
	}

	if err := h.DB.Create(&comment).Error; err != nil {
		log.Println(err)
		return &echo.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create comment.",
		}
	}

	return c.JSON(http.StatusCreated, responseFormatter[model.Comment](comment, nil, os.Getenv("DOMAIN")))
}

func (h *Handler) EditComment(c echo.Context) error {
	_, err := h.isOwnerOrAdmin(c, c.Param("id"), "comment")
	if err != nil {
		return err
	}
	id := c.Param("id")

	v := model.EditComment{}
	if err := c.Bind(&v); err != nil {
		log.Println(err)
		return err
	}

	if err := c.Validate(&v); err != nil {
		log.Println(err)
		return err
	}

	// TODO: Sign
	r := h.DB.Model(&model.Comment{ID: id}).Update("body", v.Body)
	if r.Error != nil {
		log.Println(err)
		return &echo.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update comment.",
		}
	}

	return c.JSON(http.StatusOK, UpdateResponse{Updated: 1})
}

func (h *Handler) DeleteComment(c echo.Context) error {
	_, err := h.isOwnerOrAdmin(c, c.Param("id"), "comment")
	if err != nil {
		return err
	}
	id := c.Param("id")

	r := h.DB.Delete(&model.Comment{ID: id})
	if r.Error != nil {
		log.Println(err)
		return &echo.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete comment.",
		}
	}

	if r.RowsAffected == 0 {
		return &echo.HTTPError{
			Code:    http.StatusNotFound,
			Message: "Comment not found.",
		}
	}

	return c.JSON(http.StatusOK, DeleteResponse{Deleted: 1})
}
