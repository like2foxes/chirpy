package api

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/like2foxes/chirpy/internal/database"
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

func GetChirps(w http.ResponseWriter, r *http.Request) {
	db, err := database.NewDB("database.json")
	if err != nil {
		log.Println("Error creating db")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	chirps, err := db.GetChirps()
	if err != nil {
		log.Println("Error getting chirps")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, chirps)
}

func GetChirp(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	db, err := database.NewDB("database.json")
	if err != nil {
		log.Println("Error creating db")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	idAsInt, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Error converting id to int")
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	chirp, err := db.GetChirp(idAsInt)
	if err != nil {
		log.Println("Error getting chirp")
		if err.Error() == "not found" {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, chirp)
}

func PostChirp(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var c chirp
	err := decoder.Decode(&c)

	if err != nil {
		log.Println("Error decoding chirp")
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if len(c.Body) > 140 {
		log.Println("Chirp is too long")
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleaned := cleanData(c.Body)
	db, err := database.NewDB("database.json")
	if err != nil {
		log.Println("Error creating db")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	chirp, err := db.CreateChirp(cleaned)
	if err != nil {
		log.Println("Error creating chirp")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, chirp)
}
