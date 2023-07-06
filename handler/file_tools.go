package handler

import (
	"tbd/model"
)

func modelPublicFile(f model.File) model.PublicFile {
	return model.PublicFile{
		ID:    f.ID,
		Title: f.Title,
		Path:  f.Path,
	}
}

func modelPublicFiles(f []model.File) []model.PublicFile {
	var publicFiles []model.PublicFile
	for _, v := range f {
		publicFiles = append(publicFiles, modelPublicFile(v))
	}
	return publicFiles
}
