package repository

import (
	"time"

	"Avito-test-task/gen/api"
)

type User struct {
	UserId   string `gorm:"primaryKey"`
	Username string
	IsActive bool
}

type Team struct {
	TeamName string `gorm:"primaryKey"`
	Members  []User `gorm:"many2many:team_members;joinForeignKey:TeamName;joinReferences:UserId"`
}

type PullRequest struct {
	PullRequestId     string `gorm:"primaryKey"`
	PullRequestName   string
	AuthorId          string
	Status            api.PullRequestStatus
	CreatedAt         *time.Time
	MergedAt          *time.Time
	AssignedReviewers []User `gorm:"many2many:pull_request_reviewers;joinForeignKey:PullRequestId;joinReferences:UserId"`
}
