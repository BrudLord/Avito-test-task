package main

import (
	"Avito-test-task/internal/repository"
	"Avito-test-task/internal/server"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"time"

	"Avito-test-task/gen/api"
	"github.com/go-chi/chi/v5"
)

func dbInit(db *gorm.DB) {
	// Создаем пользователей
	user1 := repository.User{UserId: "u1", Username: "Alice", IsActive: true}
	user2 := repository.User{UserId: "u2", Username: "Bob", IsActive: true}
	user3 := repository.User{UserId: "u3", Username: "Charlie", IsActive: true}
	user4 := repository.User{UserId: "u4", Username: "Dave", IsActive: true}
	user5 := repository.User{UserId: "u5", Username: "Eve", IsActive: true}

	db.Create(&user1)
	db.Create(&user2)
	db.Create(&user3)
	db.Create(&user4)
	db.Create(&user5)

	// Создаем команды
	team1 := repository.Team{
		TeamName: "TeamAlpha",
		Members:  []repository.User{user1, user2, user3, user4},
	}
	team2 := repository.Team{
		TeamName: "TeamBeta",
		Members:  []repository.User{user5},
	}

	db.Create(&team1)
	db.Create(&team2)

	// Время для пул реквеста
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	// Создаем пул реквесты
	pr1 := repository.PullRequest{
		PullRequestId:     "pr1",
		PullRequestName:   "Add feature X",
		AuthorId:          user1.UserId,
		Status:            api.PullRequestStatusMERGED, // закрытый
		CreatedAt:         &yesterday,
		MergedAt:          &now,
		AssignedReviewers: []repository.User{user2, user3},
	}

	pr2 := repository.PullRequest{
		PullRequestId:     "pr2",
		PullRequestName:   "Fix bug Y",
		AuthorId:          user2.UserId,
		Status:            api.PullRequestStatusOPEN, // открытый
		CreatedAt:         &now,
		AssignedReviewers: []repository.User{user1},
	}

	pr3 := repository.PullRequest{
		PullRequestId:     "pr3",
		PullRequestName:   "Update docs",
		AuthorId:          user3.UserId,
		Status:            api.PullRequestStatusOPEN, // открытый
		CreatedAt:         &now,
		AssignedReviewers: []repository.User{user4},
	}

	db.Create(&pr1)
	db.Create(&pr2)
	db.Create(&pr3)
}

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=db port=8989 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	r := chi.NewRouter()

	repo := repository.New(db)
	repo.Init()
	dbInit(db)

	srv := server.NewServer(repo)

	handler := api.Handler(srv)
	r.Mount("/", handler)

	log.Println("Server started on :8080")
	http.ListenAndServe("0.0.0.0:8080", r)
}
