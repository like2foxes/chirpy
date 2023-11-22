package main

import (
	"flag"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/like2foxes/chirpy/internal/api"
	"github.com/like2foxes/chirpy/internal/database"
	"log"
	"net/http"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fileRoot := os.Getenv("FILE_ROOT")
	port := os.Getenv("PORT")
	databaseFile := os.Getenv("DATABASE_FILE")
	jwtSecret := os.Getenv("JWT_SECRET")

	dbg := flag.Bool("debug", false, "enable debug mode")
	flag.Parse()
	if *dbg {
		onDebug(databaseFile)
	}

	db, err := database.NewDB(databaseFile)
	if err != nil {
		log.Fatal(err)
	}

	apiCfg := api.NewApiConfig(jwtSecret, db, 0)
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
	apiRouter.Get("/chirps", apiCfg.GetChirps)
	apiRouter.Get("/chirps/{id}", apiCfg.GetChirp)
	apiRouter.Post("/chirps", apiCfg.PostChirp)
	apiRouter.Get("/users", apiCfg.GetUsers)
	apiRouter.Get("/users/{id}", apiCfg.GetUser)
	apiRouter.Post("/users", apiCfg.PostUser)
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

func onDebug(databaseFile string) {
	log.Println("Debug mode enabled")
	err := os.Remove(databaseFile)
	if err != nil {
		log.Println(err)
	}
}
