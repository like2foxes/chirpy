package main

import (
	"flag"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/like2foxes/chirpy/internal/api"
	"log"
	"net/http"
	"os"
)

func main() {
	const fileRoot = "./public"
	const port = "8080"

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	dbg := flag.Bool("debug", false, "enable debug mode")
	flag.Parse()
	if *dbg {
		log.Println("Debug mode enabled")
		err := os.Remove("database.json")
		if err != nil {
			log.Println(err)
		}
	}

	apiCfg := api.NewApiConfig()
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
	apiRouter.Get("/healthz", api.GetHealthz)
	apiRouter.Get("/reset", apiCfg.GetReset)
	apiRouter.Get("/chirps", api.GetChirps)
	apiRouter.Get("/chirps/{id}", api.GetChirp)
	apiRouter.Post("/chirps", api.PostChirp)
	apiRouter.Get("/users", api.GetUsers)
	apiRouter.Get("/users/{id}", api.GetUser)
	apiRouter.Post("/users", api.PostUser)
	apiRouter.Post("/login", apiCfg.PostLogin)
	apiRouter.Put("/users", apiCfg.PutUser)

	r.Mount("/admin", adminRouter)
	adminRouter.Get("/metrics", apiCfg.GetMetrics)

	corsMux := api.MiddlewareCors(r)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}
	log.Println("Server started at port " + port)
	log.Fatal(server.ListenAndServe())
}
