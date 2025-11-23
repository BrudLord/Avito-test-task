package server

import (
	"Avito-test-task/gen/api"
	"Avito-test-task/internal/repository"
	"Avito-test-task/internal/wrappers"
	"encoding/json"
	"net/http"
)

type Server struct {
	repo repository.Repository
}

func NewServer(repo repository.Repository) *Server {
	return &Server{repo: repo}
}

func writeErrorResponse(w http.ResponseWriter, status int, code api.ErrorResponseErrorCode, message string) {
	var resp = api.ErrorResponse{}
	resp.Error.Code = code
	resp.Error.Message = message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, "Can't send ErrorResponse", status)
		return
	}
}

// PostPullRequestMerge Пометить PR как MERGED (идемпотентная операция)
func (s *Server) PostPullRequestMerge(w http.ResponseWriter, r *http.Request) {
	var req api.PullRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if hasErr, err := s.repo.PullRequestMerge(&req); hasErr {
		if err == api.NOTFOUND {
			writeErrorResponse(w, http.StatusNotFound, api.NOTFOUND, "resource not found")
		}
		return
	}
	var resp = wrappers.PrWrapper{PullRequest: req}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, "Can't send Ok Response", http.StatusInternalServerError)
		return
	}
}

// PostPullRequestReassign Переназначить конкретного ревьювера на другого из его команды
func (s *Server) PostPullRequestReassign(w http.ResponseWriter, r *http.Request) {
	var req wrappers.UserSwitch
	var resp = wrappers.SwitchPrWrapper{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if hasErr, err := s.repo.PullRequestReassign(&req, &resp); hasErr {
		if err == api.NOTFOUND {
			writeErrorResponse(w, http.StatusNotFound, api.NOTFOUND, "resource not found")
		} else if err == api.PRMERGED {
			writeErrorResponse(w, http.StatusConflict, api.PRMERGED, "cannot reassign on merged PR")
		} else if err == api.NOTASSIGNED {
			writeErrorResponse(w, http.StatusConflict, api.NOTASSIGNED, "reviewer is not assigned to this PR")
		} else if err == api.NOCANDIDATE {
			writeErrorResponse(w, http.StatusConflict, api.NOCANDIDATE, "no active replacement candidate in team")
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, "Can't send Ok Response", http.StatusInternalServerError)
		return
	}
}

// PostTeamAdd Создать команду с участниками (создаёт/обновляет пользователей)
func (s *Server) PostTeamAdd(w http.ResponseWriter, r *http.Request) {
	var req api.Team
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if hasErr, err := s.repo.CreateTeam(&req); hasErr {
		if err == api.TEAMEXISTS {
			writeErrorResponse(w, http.StatusBadRequest, api.TEAMEXISTS, req.TeamName+" already exists")
		}
		return
	}

	var resp = wrappers.TeamWrapper{Team: req}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, "Can't send Ok Response", http.StatusInternalServerError)
		return
	}
}

// GetTeamGet Получить команду с участниками
func (s *Server) GetTeamGet(w http.ResponseWriter, r *http.Request, params api.GetTeamGetParams) {
	var resp = api.Team{TeamName: params.TeamName}

	if hasErr, err := s.repo.GetTeam(&resp); hasErr {
		if err == api.NOTFOUND {
			writeErrorResponse(w, http.StatusBadRequest, api.NOTFOUND, "resource not found")
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, "Can't send Ok Response", http.StatusInternalServerError)
		return
	}
}

// GetUsersGetReview Получить PR'ы, где пользователь назначен ревьювером
func (s *Server) GetUsersGetReview(w http.ResponseWriter, r *http.Request, params api.GetUsersGetReviewParams) {
	var resp = wrappers.UserPRs{ID: params.UserId}

	if hasErr, _ := s.repo.UserPRs(&resp); hasErr {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, "Can't send Ok Response", http.StatusInternalServerError)
		return
	}
}

// PostUsersSetIsActive Установить флаг активности пользователя
func (s *Server) PostUsersSetIsActive(w http.ResponseWriter, r *http.Request) {
	var req api.User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if hasErr, err := s.repo.UpdateUser(&req); hasErr {
		if err == api.NOTFOUND {
			writeErrorResponse(w, http.StatusNotFound, api.NOTFOUND, "resource not found")
		}
		return
	}

	var resp = wrappers.UserWrapper{User: req}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, "Can't send Ok Response", http.StatusInternalServerError)
		return
	}
}

// PostPullRequestCreate Создать PR и автоматически назначить до 2 ревьюверов из команды автора
func (s *Server) PostPullRequestCreate(w http.ResponseWriter, r *http.Request) {
	var req api.PullRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if hasErr, err := s.repo.MakePR(&req); hasErr {
		if err == api.NOTFOUND {
			writeErrorResponse(w, http.StatusNotFound, api.NOTFOUND, "resource not found")
		} else if err == api.PREXISTS {
			writeErrorResponse(w, http.StatusConflict, api.PREXISTS, "PR id already exists")
		}
		return
	}

	var resp = wrappers.PrWrapper{PullRequest: req}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, "Can't send Ok Response", http.StatusInternalServerError)
		return
	}
}
