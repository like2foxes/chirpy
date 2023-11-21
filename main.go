package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

func main() {
	const fileRoot = "."
	const port = "8080"

	dbg := flag.Bool("debug", false, "enable debug mode")
	flag.Parse()
	if *dbg {
		log.Println("Debug mode enabled")
		err := os.Remove("database.json")
		if err != nil {
			log.Println(err)
		}
	}

	apiCfg := &apiConfig{fileserverHits: 0}
	fsHandler := apiCfg.MiddlewareMetricsInc(
		http.StripPrefix(
			"/app",
			http.FileServer(http.Dir(fileRoot))),
	)

	r := chi.NewRouter()
	apiRouter := chi.NewRouter()
	adminRouter := chi.NewRouter()

	r.Handle("/app/*", fsHandler)
	r.Handle("/app", fsHandler)

	r.Mount("/api", apiRouter)
	apiRouter.Get("/healthz", getHealthz)
	apiRouter.Get("/reset", apiCfg.getReset)
	apiRouter.Get("/chirps", getChirps)
	apiRouter.Get("/chirps/{id}", getChirp)
	apiRouter.Post("/chirps", postChirp)
	apiRouter.Get("/users", getUsers)
	apiRouter.Get("/users/{id}", getUser)
	apiRouter.Post("/users", postUser)

	r.Mount("/admin", adminRouter)
	adminRouter.Get("/metrics", apiCfg.getMetrics)

	corsMux := middlewareCors(r)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}
	log.Println("Server started at port " + port)
	log.Fatal(server.ListenAndServe())
}
