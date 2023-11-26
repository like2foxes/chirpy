package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (c *ApiConfig) PostLogin(w http.ResponseWriter, r *http.Request) {
	var u user
	if !decodeItemOr404(w, r, &u) {
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

	accessToken, err := c.createJWTTokenForUser(user.Id, time.Now().Add(time.Hour), "chirpy-access")
	if err != nil {
		internalServerError(w, err)
		return
	}

	refreshToken, err := c.createJWTTokenForUser(user.Id, time.Now().Add(time.Hour*24*60), "chirpy-refresh")
	if err != nil {
		internalServerError(w, err)
		return
	}

	postLoginResponse := postLoginResponse{
		Id:           user.Id,
		Email:        user.Email,
		Token:        accessToken,
		RefreshToken: refreshToken,
	}
	respondWithJSON(w, http.StatusOK, postLoginResponse)
}

func (c ApiConfig) PostRefresh(w http.ResponseWriter, r *http.Request) {
	claims, ok := parseClaimsFromHeader(w, r, c.jwtSecret)
	if !ok {
		return
	}
	if claims.Issuer != "chirpy-refresh" {
		tokenParsingError(w, errors.New("invalid token issuer"))
		return
	}

	token, ok := getTokenStringFromHeader(w, r)
	if !ok {
		tokenParsingError(w, errors.New("invalid token"))
		return
	}

	if c.db.IsTokenRevoked(token) {
		tokenParsingError(w, errors.New("token is revoked"))
		return
	}

	id, err := strconv.Atoi(claims.Subject)
	if err != nil {
		tokenParsingError(w, errors.New("invalid token"))
		return
	}

	token, err = c.createJWTTokenForUser(id, time.Now().Add(time.Hour), "chirpy-access")
	if err != nil {
		internalServerError(w, err)
		return
	}

	tokenResponse := TokenResponse{
		Token: token,
	}

	respondWithJSON(w, http.StatusOK, tokenResponse)
}

func (c ApiConfig) PostRevoke(w http.ResponseWriter, r *http.Request) {
	token, ok := getTokenStringFromHeader(w, r)
	if !ok {
		tokenParsingError(w, errors.New("invalid token"))
		return
	}

	err := c.db.RevokeToken(token)
	if err != nil {
		internalServerError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, nil)
}

type postLoginResponse struct {
	Id           int    `json:"id"`
	Email        string `json:"email"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenResponse struct {
	Token string `json:"token"`
}
