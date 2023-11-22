package api

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

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

func decodeItemOr404(w http.ResponseWriter, r *http.Request, item interface{}) bool {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&item)
	if err != nil {
		decodingError(w, err)
		return false
	}
	return true
}

func idFromURL(w http.ResponseWriter, r *http.Request) (int, bool) {
	id := chi.URLParam(r, "id")
	idAsInt, err := strconv.Atoi(id)
	if err != nil {
		conversionError(w, err)
		return 0, false
	}
	return idAsInt, true
}
