package api
import (
	"log"
	"net/http"
)

func queryError(w http.ResponseWriter, err error) {
	log.Printf("Error querying database: %s\n", err.Error())
	if err.Error() == "not found" {
		respondWithError(w, http.StatusNotFound, "not found")
		return
	}
	log.Printf("Error querying database: %s\n", err.Error())
	respondWithError(w, http.StatusInternalServerError, "database error")
}

func autherizationHeaderError(w http.ResponseWriter, err error) {
	log.Printf("Error getting authorization header: %s\n", err.Error())
	respondWithError(w, http.StatusUnauthorized, "invalid credentials")
}

func tokenParsingError(w http.ResponseWriter, err error) {
	log.Printf("Error parsing token: %s\n", err.Error())
	respondWithError(w, http.StatusUnauthorized, "invalid credentials")
}

func internalServerError(w http.ResponseWriter, err error) {
	log.Printf("Error: %s\n", err.Error())
	respondWithError(w, http.StatusInternalServerError, "internal server error")
}
