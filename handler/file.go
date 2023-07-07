package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"tbd/model"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/labstack/echo/v4"

	"github.com/google/uuid"
)

func fileExtentionFromFileName(fileName string) (string, error) {
	// Use extention after last dot; for ex.: 'somefile.txt' -> 'txt'
	// If there's no dot, return empty string
	for i := len(fileName) - 1; i >= 0; i-- {
		if fileName[i] == '.' {
			return fileName[i+1:], nil
		}
	}
	return "", fmt.Errorf("no extention found")
}

func (h *Handler) CreateFiles(c echo.Context) error {
	u, httpErr := UserFromContextHttpError(c)
	if httpErr != nil {
		return httpErr
	}

	// Multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Failed to parse multipart form."}
	}
	files := form.File["files"]

	// Create S3 client
	// This will pickup env variables
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		log.Printf("error: %v", err)
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to configure upload client."}
	}

	client := s3.NewFromConfig(cfg)

	dbFiles := []model.File{}

	for _, file := range files {
		src, err := file.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		// Assemble new file DB record
		newID, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		dbFile := model.File{}
		dbFile.ID = newID.String()

		newFilename := ""
		fileExtention, err := fileExtentionFromFileName(file.Filename)
		if err == nil {
			newFilename = fmt.Sprintf("%s.%s", dbFile.ID, fileExtention)
		} else {
			newFilename = fmt.Sprintf("%s", dbFile.ID)
		}

		dbFile.Title = file.Filename
		dbFile.Path = fmt.Sprintf("%s%s", "general/", newFilename)
		dbFile.Mime = file.Header.Get("Content-Type")
		dbFile.Size = file.Size
		dbFile.CreatedBy = u
		dbFile.IsProvisional = true

		// Upload file
		uploader := manager.NewUploader(client)
		bucket := os.Getenv("AWS_BUCKET_NAME")

		result, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(dbFile.Path),
			Body:   src,
		})

		if err != nil {
			log.Printf("Failed to upload file: %v", err)
			return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to upload file."}
		}

		// Save file to DB
		r := h.DB.Create(&dbFile)
		if r.Error != nil {
			log.Printf("Failed to save file to DB: %v", r.Error)
			return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to save file to DB"}
		}

		fmt.Sprintln("File uploaded to", result.Location)
		dbFiles = append(dbFiles, dbFile)
	}

	return c.JSON(http.StatusCreated, struct {
		Files []model.File `json:"files"`
	}{Files: dbFiles})
}

func (h *Handler) GetFiles(c echo.Context) error {
	files := []model.File{}
	r := h.DB.Find(&files)
	if r.Error != nil {
		log.Printf("Failed to get files from DB: %v", r.Error)
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to get files from DB"}
	}

	return c.JSON(http.StatusOK, struct {
		Files []model.File `json:"files"`
	}{Files: files})
}

func (h *Handler) DeleteFile(c echo.Context) error {
	fileID := c.Param("id")

	// Get file from DB
	file := model.File{}
	r := h.DB.Where("id = ?", fileID).First(&file)
	if r.Error != nil {
		log.Printf("Failed to get file from DB: %v", r.Error)
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to get file from DB"}
	}

	// Delete file from S3
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		log.Printf("error: %v", err)
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to configure client."}
	}

	client := s3.NewFromConfig(cfg)

	bucket := os.Getenv("AWS_BUCKET_NAME")
	_, err = client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(file.Path),
	})
	if err != nil {
		log.Printf("Failed to delete file from S3: %v", err)
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to delete file from S3"}
	}

	// Delete file from DB
	r = h.DB.Delete(&file)
	if r.Error != nil {
		log.Printf("Failed to delete file from DB: %v", r.Error)
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to delete file from DB"}
	}

	return c.JSON(http.StatusOK, DeleteResponse{Deleted: r.RowsAffected})
}

func (h *Handler) markFilesAsProvisioned(files []model.File) error {
	for _, file := range files {
		file.IsProvisional = false

		err := h.DB.Model(model.File{ID: file.ID}).Update("is_provisional", false).Error
		if err != nil {
			return err
		}
	}

	return nil
}
