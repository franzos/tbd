package handler

import (
	"tbd/model"
)

func validateEntryType(t string) bool {
	for _, v := range entryTypes {
		if v == t {
			return true
		}
	}
	return false
}

func modelPublicEntry(e model.Entry) model.PublicEntry {
	return model.PublicEntry{
		ID:        e.ID,
		Type:      e.Type,
		Data:      e.Data,
		Files:     modelPublicFiles(e.Files),
		CreatedBy: modelPublicUser(e.CreatedBy),
		CreatedAt: e.CreatedAt,
	}
}

func modelPublicEntries(e []model.Entry) []model.PublicEntry {
	var publicEntries []model.PublicEntry
	for _, v := range e {
		publicEntries = append(publicEntries, modelPublicEntry(v))
	}
	return publicEntries
}
