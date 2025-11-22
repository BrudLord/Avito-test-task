package server

import (
	"Avito-test-task/gen/api"
	"Avito-test-task/internal/repository"
	"net/http"
)

type Server struct {
	repo repository.Repository
}

func NewServer(repo repository.Repository) *Server {
	return &Server{repo: repo}
}

// PostPullRequestMerge Пометить PR как MERGED (идемпотентная операция)
func (s *Server) PostPullRequestMerge(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

// PostPullRequestReassign Переназначить конкретного ревьювера на другого из его команды
func (s *Server) PostPullRequestReassign(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

// PostTeamAdd Создать команду с участниками (создаёт/обновляет пользователей)
func (s *Server) PostTeamAdd(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

// GetTeamGet Получить команду с участниками
func (s *Server) GetTeamGet(w http.ResponseWriter, r *http.Request, params api.GetTeamGetParams) {
	//TODO implement me
	panic("implement me")
}

// GetUsersGetReview Получить PR'ы, где пользователь назначен ревьювером
func (s *Server) GetUsersGetReview(w http.ResponseWriter, r *http.Request, params api.GetUsersGetReviewParams) {
	//TODO implement me
	panic("implement me")
}

// PostUsersSetIsActive Установить флаг активности пользователя
func (s *Server) PostUsersSetIsActive(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

// PostPullRequestCreate Создать PR и автоматически назначить до 2 ревьюверов из команды автора
func (s *Server) PostPullRequestCreate(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}
