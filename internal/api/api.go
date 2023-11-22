package api

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
)

func GetHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func respondWithError(w http.ResponseWriter, status int, msg string) {
	log.Println(msg)
	ce := chirpError{Error: msg}
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")
	error, _ := json.Marshal(ce)
	w.Write(error)
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

func cleanData(body string) string {
	profane := []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}

	var words []string
	for _, word := range strings.Split(body, " ") {
		if slices.Contains(profane, strings.ToLower(word)) {
			words = append(words, "****")
		} else {
			words = append(words, word)
		}
	}

	return strings.Join(words, " ")
}
