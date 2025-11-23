package main

import (
	"Avito-test-task/gen/api"
	"Avito-test-task/internal/repository"
	"Avito-test-task/internal/server"
	"fmt"
	"github.com/go-chi/chi/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
)

func main() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	conStr := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port,
	)

	db, err := gorm.Open(postgres.Open(conStr), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	r := chi.NewRouter()

	repo := repository.New(db)
	repo.Init()

	srv := server.NewServer(repo)

	handler := api.Handler(srv)
	r.Mount("/", handler)

	log.Println("Server started on :8080")
	http.ListenAndServe("0.0.0.0:8080", r)
}
