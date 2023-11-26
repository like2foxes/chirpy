package api

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/like2foxes/chirpy/internal/database"
)

type user struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds *int   `json:"expires_in_seconds,omitempty"`
}

type noPasswordUser struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

func (c ApiConfig) PutUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := parseClaimsFromHeader(w, r, c.jwtSecret)
	if !ok {
		return
	}

	log.Printf("claims: %+v\n", claims)
	if !isIssuerIsAccess(claims) {
		tokenParsingError(w, errors.New("invalid token issuer"))
		return
	}

	if !validateExpirationTime(w, r, claims) {
		return
	}

	var u user
	if !decodeItemOr404(w, r, &u) {
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
	var u user
	if !decodeItemOr404(w, r, &u) {
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
	id, ok := idFromURL(w, r)
	if !ok {
		return
	}
	user, err := c.db.GetUser(id)
	if err != nil {
		queryError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, user)
}

func newNoPasswordUser(u database.User) noPasswordUser {
	return noPasswordUser{
		Id:    u.Id,
		Email: u.Email,
	}
}

func (c ApiConfig) createJWTTokenForUser(id int, exTime time.Time, issuer string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{})
	claims := jwt.RegisteredClaims{
		Issuer:   issuer,
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(exTime),
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

func parseClaimsFromHeader(w http.ResponseWriter, r *http.Request, secret string) (jwt.RegisteredClaims, bool) {
	tokenString, ok := getTokenStringFromHeader(w, r)
	if !ok {
		return jwt.RegisteredClaims{}, false
	}
	claimes, ok := parseTokenString(w, tokenString, secret)
	if !ok {
		return jwt.RegisteredClaims{}, false
	}
	return *claimes, true
}

func parseTokenString(
	w http.ResponseWriter,
	tokenString string,
	secret string,
) (*jwt.RegisteredClaims, bool) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		},
	)
	if err != nil {
		tokenParsingError(w, err)
		return nil, false
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		tokenParsingError(w, errors.New("invalid token claims"))
		return nil, false
	}
	return claims, true
}

func getTokenStringFromHeader(w http.ResponseWriter, r *http.Request) (string, bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		autherizationHeaderError(w, errors.New("no authorization header"))
		return "", false
	}
	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 {
		autherizationHeaderError(w, errors.New("invalid authorization header"))
		return "", false
	}
	tokenString := authHeaderParts[1]
	return tokenString, true
}

func isIssuerIsAccess(claims jwt.RegisteredClaims) bool {
	return claims.Issuer == "chirpy-access"
}

func validateExpirationTime(w http.ResponseWriter, r *http.Request, claims jwt.RegisteredClaims) bool {
	expiresAt, err := claims.GetExpirationTime()
	if err != nil {
		tokenParsingError(w, err)
		return false
	}
	if jwt.NewNumericDate(time.Now().UTC()).After(expiresAt.Time) {
		tokenParsingError(w, errors.New("token expired"))
		return false
	}
	return true
}
