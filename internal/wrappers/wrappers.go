package wrappers

import "Avito-test-task/gen/api"

type TeamWrapper struct {
	Team api.Team `json:"team"`
}

type UserWrapper struct {
	User api.User `json:"user"`
}

type PrWrapper struct {
	PullRequest api.PullRequest `json:"pr"`
}

type UserSwitch struct {
	PullRequest string `json:"pull_request_id"`
	User        string `json:"old_reviewer_id"`
}

type SwitchPrWrapper struct {
	PullRequest api.PullRequest `json:"pr"`
	ReplacedBy  string          `json:"replaced_by"`
}

type UserPRs struct {
	ID           string                 `json:"user_id"`
	PullRequests []api.PullRequestShort `json:"pull_requests"`
}
