package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"rehab-app/internal/api"
	"rehab-app/internal/config"
	"rehab-app/internal/db"
	appmiddleware "rehab-app/internal/middleware"
)

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.RunMigrations {
		if err := db.RunMigrations(database, "migrations"); err != nil {
			log.Fatal(err)
		}
	}

	if cfg.SeedData {
		if err := db.Seed(database); err != nil {
			log.Fatal(err)
		}
	}

	router := chi.NewRouter()
	router.Use(middleware.RealIP)
	router.Use(appmiddleware.Logger)
	router.Use(appmiddleware.Recover)

	apiHandler := api.New(database, cfg)
	router.Mount("/api/v1", apiHandler.Router())

	spaDir := "build"
	fileServer := http.FileServer(http.Dir(spaDir))

	router.Handle("/assets/*", fileServer)
	router.Handle("/favicon.ico", fileServer)
	router.Handle("/manifest.json", fileServer)

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		path := filepath.Join(spaDir, filepath.Clean(r.URL.Path))
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}

		http.ServeFile(w, r, filepath.Join(spaDir, "index.html"))
	})

	log.Printf("Server running on %s", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, router))
}
