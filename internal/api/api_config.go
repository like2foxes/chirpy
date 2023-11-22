package api

import (
	"fmt"
	"net/http"
	"os"
)

type ApiConfig struct {
	fileserverHits int
	jwtSecret      string
}

func NewApiConfig() *ApiConfig {
	jwtSecret := os.Getenv("JWT_SECRET")
	return &ApiConfig{fileserverHits: 0, jwtSecret: jwtSecret}
}

func (c *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (c *ApiConfig) GetMetrics(w http.ResponseWriter, r *http.Request) {
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

func (c *ApiConfig) GetReset(w http.ResponseWriter, r *http.Request) {
	c.fileserverHits = 0
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits: " + fmt.Sprintf("%d", c.fileserverHits)))
}
