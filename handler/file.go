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

func (h *Handler) CreateFiles(c echo.Context) (err error) {
	u, err := userFromToken(c)
	if err != nil {
		log.Printf("error: %v", err)
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to parse provided token."}
	}

	// Multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	files := form.File["files"]

	// Create S3 client
	// This will pickup env variables
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		log.Printf("error: %v", err)
		return
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
