package api

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
)

type chirp struct {
	Body string `json:"body"`
}

type chirpError struct {
	Error string `json:"error"`
}

type chirpValid struct {
	Cleansed string `json:"cleaned_body"`
}

func (c ApiConfig) GetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := c.db.GetChirps()
	if err != nil {
		log.Println("Error getting chirps")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, chirps)
}

func (c ApiConfig) GetChirp(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idAsInt, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Error converting id to int")
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	chirp, err := c.db.GetChirp(idAsInt)
	if err != nil {
		queryError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, chirp)
}

func (c ApiConfig) PostChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var ch chirp
	err := decoder.Decode(&c)

	if err != nil {
		log.Println("Error decoding chirp")
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if len(ch.Body) > 140 {
		log.Println("Chirp is too long")
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleaned := cleanData(ch.Body)

	chirp, err := c.db.CreateChirp(cleaned)
	if err != nil {
		log.Println("Error creating chirp")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, chirp)
}
