package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/like2foxes/chripy/internal/database"
)

type apiConfig struct {
	fileserverHits int
}

func (c *apiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (c *apiConfig) getMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(
		"<html><body>" +
			"<h1>Welcome, Chirpy Admin</h1>" +
			"<p>" +
			fmt.Sprintf("Chirpy has been visited %d times!", c.fileserverHits) +
			"</p>" +
			"</body></html>"))
}

func (c *apiConfig) getReset(w http.ResponseWriter, r *http.Request) {
	c.fileserverHits = 0
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits: " + fmt.Sprintf("%d", c.fileserverHits)))
}

func getHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func getChirps(w http.ResponseWriter, r *http.Request) {
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

func getChirp(w http.ResponseWriter, r *http.Request) {
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

func postChirp(w http.ResponseWriter, r *http.Request) {

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

func postUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var u user
	err := decoder.Decode(&u)
	if err != nil {
		log.Println("Error decoding user")
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	db, err := database.NewDB("database.json")
	if err != nil {
		log.Println("Error creating db")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	user, err := db.CreateUser(u.Email, u.Password)
	if err != nil {
		log.Println("Error creating user")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	noPWUser := newNoPasswordUser(user)
	
	respondWithJSON(w, http.StatusCreated, noPWUser)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	db, err := database.NewDB("database.json")
	if err != nil {
		log.Println("Error creating db")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	users, err := db.GetUsers()
	if err != nil {
		log.Println("Error getting users")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, users)
}

func getUser(w http.ResponseWriter, r *http.Request) {
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
	user, err := db.GetUser(idAsInt)
	if err != nil {
		log.Println("Error getting user")
		if err.Error() == "not found" {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, user)
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

type chirp struct {
	Body string `json:"body"`
}

type user struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type chirpError struct {
	Error string `json:"error"`
}

type chirpValid struct {
	Cleansed string `json:"cleaned_body"`
}

type noPasswordUser struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

func newNoPasswordUser(u database.User) noPasswordUser {
	return noPasswordUser{
		Id:    u.Id,
		Email: u.Email,
	}
}
