package api

import (
	"errors"
	"net/http"
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
		queryError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, chirps)
}

func (c ApiConfig) GetChirp(w http.ResponseWriter, r *http.Request) {
	if id, ok := idFromURL(w, r); ok {
		chirp, err := c.db.GetChirp(id)
		if err != nil {
			queryError(w, err)
			return
		}
		respondWithJSON(w, http.StatusOK, chirp)
	}
}

func (c ApiConfig) PostChirp(w http.ResponseWriter, r *http.Request) {
	var ch chirp
	if !decodeItemOr404(w, r, &ch) {
		return
	}

	if len(ch.Body) > 140 {
		chirpLengthError(w, errors.New("Chirp is too long"))
		return
	}

	cleaned := cleanData(ch.Body)

	newChrip, err := c.db.CreateChirp(cleaned)
	if err != nil {
		queryError(w, err)
		return
	}
	respondWithJSON(w, http.StatusCreated, newChrip)
}
