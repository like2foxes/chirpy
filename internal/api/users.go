package api

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/like2foxes/chirpy/internal/database"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type user struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds *int   `json:"expires_in_seconds,omitempty"`
}

type noPasswordUser struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
	Token string `json:"token,omitempty"`
}

func (apiCfg ApiConfig) PutUser(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Println("Error getting authorization header")
		respondWithError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 {
		log.Println("Error getting authorization header")
		respondWithError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	tokenString := authHeaderParts[1]
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(apiCfg.jwtSecret), nil
		},
	)
	if err != nil {
		log.Println("Error parsing token")
		respondWithError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		log.Println("Error parsing token")
		respondWithError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	expiresAt, err := claims.GetExpirationTime()
	if err != nil {
		log.Println("Error parsing token")
		respondWithError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if jwt.NewNumericDate(time.Now().UTC()).After(expiresAt.Time) {
		log.Println("token expired")
		respondWithError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	decoder := json.NewDecoder(r.Body)
	var u user
	err = decoder.Decode(&u)
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

	id, err := strconv.Atoi(claims.Subject)
	if err != nil {
		log.Println("Error parsing token")
		respondWithError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	user, err := db.GetUser(id)
	if err != nil {
		log.Println("Error getting user")
		if err.Error() == "not found" {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updatedUser := database.User{
		Id:       user.Id,
		Email:    u.Email,
		Password: u.Password,
	}
	response, err := db.UpdateUser(updatedUser)
	if err != nil {
		log.Println("Error updating user")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	noPWUser := newNoPasswordUser(response)
	respondWithJSON(w, http.StatusOK, noPWUser)
}

func PostUser(w http.ResponseWriter, r *http.Request) {
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

func GetUsers(w http.ResponseWriter, r *http.Request) {
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

func GetUser(w http.ResponseWriter, r *http.Request) {
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

func (a *ApiConfig) PostLogin(w http.ResponseWriter, r *http.Request) {
	db, err := database.NewDB("database.json")
	if err != nil {
		log.Println("Error creating db")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	decoder := json.NewDecoder(r.Body)
	var u user
	err = decoder.Decode(&u)
	if err != nil {
		log.Println("Error decoding user")
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	user, err := db.GetUserByEmail(u.Email)
	if err != nil {
		log.Println("Error getting user by email")
		if err.Error() == "not found" {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password)) != nil {
		log.Println("Error comparing password")
		respondWithError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := a.createJWTTokenForUser(u, user.Id)
	if err != nil {
		log.Println("Error creating token")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err != nil {
		log.Println("Error updating user")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	postLoginResponse := postLoginResponse{
		Id:    user.Id,
		Email: user.Email,
		Token: token,
	}
	respondWithJSON(w, http.StatusOK, postLoginResponse)
}

type postLoginResponse struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
	Token string `json:"token"`
}

func newNoPasswordUser(u database.User) noPasswordUser {
	return noPasswordUser{
		Id:    u.Id,
		Email: u.Email,
	}
}

func getTimeForExpirecy(user user) time.Time {
	if user.ExpiresInSeconds == nil {
		return time.Now().UTC().Add(time.Hour * 24)
	}
	if *user.ExpiresInSeconds > 24 * 60 * 60 {
		return time.Now().UTC().Add(time.Hour * 24)
	}
	return time.Now().UTC().Add(
		time.Second * time.Duration(*user.ExpiresInSeconds),
	)
}
func (apiCfg ApiConfig) createJWTTokenForUser(user user, id int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{})
	claims := jwt.RegisteredClaims{
		Issuer:   "chirpy",
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(
			getTimeForExpirecy(user),
		),
		Subject: strconv.Itoa(id),
	}
	token.Claims = claims
	tokenString, err := token.SignedString([]byte(apiCfg.jwtSecret))
	if err != nil {
		log.Println(err.Error())
		log.Println("Error signing token")
		return "", err
	}
	return tokenString, nil
}
