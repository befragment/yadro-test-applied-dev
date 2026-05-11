package utils

import (
	"log"
	"net/http"
)

const (
	ErrInternalServer = "Internal server error"
	ErrNotFound       = "Not found"
)

func RespondWithError(w http.ResponseWriter, httpStatus int, message string) {
	RespondWithJSON(w, httpStatus, map[string]string{"error": message})
}

func RespondInternalServerError(w http.ResponseWriter, err error) {
	log.Printf("Internal server error: %v\n", err)
	RespondWithError(w, http.StatusInternalServerError, ErrInternalServer)
}

func RespondBadRequest(w http.ResponseWriter, message string) {
	RespondWithJSON(w, http.StatusBadRequest, map[string]string{
		"error": message,
	})
}

func RespondNotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = ErrNotFound
	}
	RespondWithJSON(w, http.StatusNotFound, map[string]string{
		"error": message,
	})
}