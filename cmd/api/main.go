package main

import (
	"Avito-test-task/internal/repository"
	"Avito-test-task/internal/server"
	"database/sql"
	"log"
	"net/http"

	"Avito-test-task/gen/api"
	"github.com/go-chi/chi/v5"
)

func main() {
	connStr := "user=username dbname=mydb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	r := chi.NewRouter()

	repo := repository.New(db)

	srv := server.NewServer(repo)

	handler := api.Handler(srv)
	r.Mount("/", handler)

	log.Println("Server started on :8080")
	http.ListenAndServe("0.0.0.0:8080", r)
}
