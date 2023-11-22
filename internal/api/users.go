package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/like2foxes/chirpy/internal/database"
	"golang.org/x/crypto/bcrypt"
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

func (c ApiConfig) PutUser(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		autherizationHeaderError(w, errors.New("no authorization header"))
		return
	}
	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 {
		autherizationHeaderError(w, errors.New("invalid authorization header"))
		return
	}
	tokenString := authHeaderParts[1]
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(c.jwtSecret), nil
		},
	)
	if err != nil {
		tokenParsingError(w, err)
		return
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		tokenParsingError(w, errors.New("invalid token claims"))
		return
	}

	expiresAt, err := claims.GetExpirationTime()
	if err != nil {
		tokenParsingError(w, err)
		return
	}
	if jwt.NewNumericDate(time.Now().UTC()).After(expiresAt.Time) {
		tokenParsingError(w, errors.New("token expired"))
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

	id, err := strconv.Atoi(claims.Subject)
	if err != nil {
		tokenParsingError(w, err)
		return
	}

	user, err := c.db.GetUser(id)
	if err != nil {
		queryError(w, err)
		return
	}
	updatedUser := database.User{
		Id:       user.Id,
		Email:    u.Email,
		Password: u.Password,
	}
	response, err := c.db.UpdateUser(updatedUser)
	if err != nil {
		queryError(w, err)
		return
	}
	noPWUser := newNoPasswordUser(response)
	respondWithJSON(w, http.StatusOK, noPWUser)
}

func (c ApiConfig) PostUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var u user
	err := decoder.Decode(&u)
	if err != nil {
		log.Println("Error decoding user")
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	user, err := c.db.CreateUser(u.Email, u.Password)
	if err != nil {
		queryError(w, err)
		return
	}
	noPWUser := newNoPasswordUser(user)

	respondWithJSON(w, http.StatusCreated, noPWUser)
}

func (c ApiConfig) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := c.db.GetUsers()
	if err != nil {
		queryError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, users)
}

func (c ApiConfig) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idAsInt, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Error converting id to int")
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	user, err := c.db.GetUser(idAsInt)
	if err != nil {
		queryError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, user)
}

func (c *ApiConfig) PostLogin(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var u user
	err := decoder.Decode(&u)
	if err != nil {
		log.Println("Error decoding user")
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	user, err := c.db.GetUserByEmail(u.Email)
	if err != nil {
		queryError(w, err)
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password)) != nil {
		autherizationHeaderError(w, errors.New("invalid password"))
		return
	}

	token, err := c.createJWTTokenForUser(u, user.Id)
	if err != nil {
		internalServerError(w, err)
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
	if *user.ExpiresInSeconds > 24*60*60 {
		return time.Now().UTC().Add(time.Hour * 24)
	}
	return time.Now().UTC().Add(
		time.Second * time.Duration(*user.ExpiresInSeconds),
	)
}
func (c ApiConfig) createJWTTokenForUser(user user, id int) (string, error) {
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
	tokenString, err := token.SignedString([]byte(c.jwtSecret))
	if err != nil {
		log.Println(err.Error())
		log.Println("Error signing token")
		return "", err
	}
	return tokenString, nil
}
