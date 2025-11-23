package repository

import (
	"Avito-test-task/gen/api"
	"Avito-test-task/internal/wrappers"
	_ "github.com/lib/pq"
	"gorm.io/gorm"
	"math/rand/v2"
	"slices"
	"time"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return Repository{db: db}
}

func (r Repository) Init() {
	r.db.AutoMigrate(
		&User{},
		&Team{},
		&PullRequest{},
	)
}

func (r Repository) PullRequestMerge(pr *api.PullRequest) (bool, api.ErrorResponseErrorCode) {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return true, ""
	}

	var prInDb PullRequest
	if err := tx.First(&prInDb, "pull_request_id = ?", pr.PullRequestId).Error; err != nil {
		tx.Rollback()
		return true, api.NOTFOUND
	}

	if prInDb.Status != api.PullRequestStatusMERGED {
		prInDb.Status = api.PullRequestStatusMERGED
		now := time.Now()
		prInDb.MergedAt = &now
		if err := tx.Save(&prInDb).Error; err != nil {
			tx.Rollback()
			return true, ""
		}
	}
	var assignedReviewers []string
	if err := tx.
		Table("pull_request_reviewers").
		Select("user_id").
		Where("pull_request_id = ?", prInDb.PullRequestId).
		Scan(&assignedReviewers).Error; err != nil {

		tx.Rollback()
		return true, ""
	}

	pr.Status = prInDb.Status
	pr.MergedAt = prInDb.MergedAt
	pr.PullRequestName = prInDb.PullRequestName
	pr.AuthorId = prInDb.AuthorId
	pr.AssignedReviewers = assignedReviewers

	if err := tx.Commit().Error; err != nil {
		return true, ""
	}
	return false, ""
}

func (r Repository) PullRequestReassign(
	userSwitch *wrappers.UserSwitch,
	switchPrWrapper *wrappers.SwitchPrWrapper,
) (bool, api.ErrorResponseErrorCode) {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return true, ""
	}

	var prInDb PullRequest
	if err := tx.First(&prInDb, "pull_request_id = ?", userSwitch.PullRequest).Error; err != nil {
		tx.Rollback()
		return true, api.NOTFOUND
	}
	if prInDb.Status == api.PullRequestStatusMERGED {
		tx.Rollback()
		return true, api.PRMERGED
	}

	var assignedReviewers []string
	if err := tx.
		Table("pull_request_reviewers").
		Select("user_id").
		Where("pull_request_id = ?", prInDb.PullRequestId).
		Scan(&assignedReviewers).Error; err != nil {

		tx.Rollback()
		return true, ""
	}

	if !slices.Contains(assignedReviewers, userSwitch.User) {
		return true, api.NOTASSIGNED
	}

	var oldUser User
	if err := tx.First(&oldUser, "user_id = ?", userSwitch.User).Error; err != nil {
		tx.Rollback()
		return true, api.NOTFOUND
	}

	var teamName string
	if err := tx.
		Table("team_members").
		Select("team_name").
		Where("user_id = ?", oldUser.UserId).
		Limit(1).
		Scan(&teamName).Error; err != nil {

		tx.Rollback()
		return true, ""
	}

	var teamUsers []string
	if err := tx.
		Table("team_members").
		Select("user_id").
		Where("team_name = ?", teamName).
		Scan(&teamUsers).Error; err != nil {

		tx.Rollback()
		return true, ""
	}

	var usedUsers = make(map[string]bool)
	for _, userID := range assignedReviewers {
		usedUsers[userID] = true
	}
	for _, userID := range teamUsers {
		var teamUser User
		if err := tx.First(&teamUser, "user_id = ?", userID).Error; err != nil {
			tx.Rollback()
			return true, ""
		}
		if !teamUser.IsActive {
			usedUsers[userID] = true
		}
	}
	usedUsers[prInDb.AuthorId] = true

	if len(usedUsers) >= len(teamUsers) {
		return true, api.NOCANDIDATE
	}

	var newUserIndex = rand.IntN(len(teamUsers) - len(usedUsers))
	var newUserId string
	for _, userId := range teamUsers {
		if usedUsers[userId] {
			continue
		}
		if newUserIndex == 0 {
			newUserId = userId
			break
		}
		newUserIndex--
	}

	if newUserId == "" {
		return true, api.NOCANDIDATE
	}

	tx.Table("pull_request_reviewers").
		Where("pull_request_id = ? AND user_id = ?", prInDb.PullRequestId, userSwitch.User).
		Delete(nil)

	tx.Table("pull_request_reviewers").
		Create(map[string]interface{}{
			"pull_request_id": prInDb.PullRequestId,
			"user_id":         newUserId,
		})

	if err := tx.
		Table("pull_request_reviewers").
		Select("user_id").
		Where("pull_request_id = ?", prInDb.PullRequestId).
		Scan(&assignedReviewers).Error; err != nil {

		tx.Rollback()
		return true, ""
	}

	switchPrWrapper.ReplacedBy = newUserId
	switchPrWrapper.PullRequest.PullRequestId = prInDb.PullRequestId
	switchPrWrapper.PullRequest.PullRequestName = prInDb.PullRequestName
	switchPrWrapper.PullRequest.Status = prInDb.Status
	switchPrWrapper.PullRequest.AuthorId = prInDb.AuthorId
	switchPrWrapper.PullRequest.AssignedReviewers = assignedReviewers

	if err := tx.Commit().Error; err != nil {
		return true, ""
	}
	return false, ""
}

func (r Repository) CreateTeam(team *api.Team) (bool, api.ErrorResponseErrorCode) {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return true, ""
	}

	var teamInDb Team
	teamInDb.TeamName = team.TeamName
	if err := tx.Create(&teamInDb).Error; err != nil {
		tx.Rollback()
		return true, api.TEAMEXISTS
	}

	for _, teamMember := range team.Members {
		var teamUser User
		teamUser.UserId = teamMember.UserId
		teamUser.IsActive = teamMember.IsActive
		teamUser.Username = teamMember.Username
		tx.Save(&teamUser)
		tx.Table("team_members").
			Create(map[string]interface{}{
				"team_name": team.TeamName,
				"user_id":   teamUser.UserId,
			})

	}

	if err := tx.Commit().Error; err != nil {
		return true, ""
	}
	return false, ""
}

func (r Repository) GetTeam(team *api.Team) (bool, api.ErrorResponseErrorCode) {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return true, ""
	}

	var teamUsers []string
	if err := tx.
		Table("team_members").
		Select("user_id").
		Where("team_name = ?", team.TeamName).
		Scan(&teamUsers).Error; err != nil {

		tx.Rollback()
		return true, api.NOTFOUND
	}
	for _, userID := range teamUsers {
		var teamUser User
		if err := tx.First(&teamUser, "user_id = ?", userID).Error; err != nil {
			tx.Rollback()
			return true, ""
		}
		var teamMember api.TeamMember
		teamMember.IsActive = teamUser.IsActive
		teamMember.UserId = teamUser.UserId
		teamMember.Username = teamUser.Username
		team.Members = append(team.Members, teamMember)
	}

	if err := tx.Commit().Error; err != nil {
		return true, ""
	}
	return false, ""
}

func (r Repository) UserPRs(userPRs *wrappers.UserPRs) (bool, api.ErrorResponseErrorCode) {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return true, ""
	}

	var pullRequestsId []string
	if err := tx.
		Table("pull_request_reviewers").
		Select("pull_request_id").
		Where("user_id = ?", userPRs.ID).
		Scan(&pullRequestsId).Error; err != nil {

		tx.Rollback()
		return true, ""
	}

	for _, pullRequestID := range pullRequestsId {
		var prInDb PullRequest
		if err := tx.First(&prInDb, "pull_request_id = ?", pullRequestID).Error; err != nil {
			tx.Rollback()
			return true, api.NOTFOUND
		}
		var shortPR api.PullRequestShort
		shortPR.PullRequestId = prInDb.PullRequestId
		shortPR.PullRequestName = prInDb.PullRequestName
		if prInDb.Status == api.PullRequestStatusMERGED {
			shortPR.Status = api.PullRequestShortStatusMERGED
		} else {
			shortPR.Status = api.PullRequestShortStatusOPEN
		}
		shortPR.AuthorId = prInDb.AuthorId
		userPRs.PullRequests = append(userPRs.PullRequests, shortPR)

	}

	if err := tx.Commit().Error; err != nil {
		return true, ""
	}
	return false, ""
}

func (r Repository) UpdateUser(user *api.User) (bool, api.ErrorResponseErrorCode) {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return true, ""
	}

	var userDb User
	if err := tx.First(&userDb, "user_id = ?", user.UserId).Error; err != nil {
		tx.Rollback()
		return true, api.NOTFOUND
	}

	userDb.IsActive = user.IsActive
	if err := tx.Save(&userDb).Error; err != nil {
		tx.Rollback()
		return true, ""
	}

	var teamName string
	if err := tx.
		Table("team_members").
		Select("team_name").
		Where("user_id = ?", userDb.UserId).
		Limit(1).
		Scan(&teamName).Error; err != nil {

		tx.Rollback()
		return true, ""
	}

	user.TeamName = teamName
	user.Username = userDb.Username

	if err := tx.Commit().Error; err != nil {
		return true, ""
	}
	return false, ""
}

func (r Repository) MakePR(pullRequest *api.PullRequest) (bool, api.ErrorResponseErrorCode) {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return true, ""
	}

	var prInDb PullRequest
	if err := tx.First(&prInDb, "pull_request_id = ?", pullRequest.PullRequestId).Error; err == nil {
		tx.Rollback()
		return true, api.PREXISTS
	}

	prInDb.PullRequestId = pullRequest.PullRequestId
	prInDb.PullRequestName = pullRequest.PullRequestName
	prInDb.Status = api.PullRequestStatusOPEN
	prInDb.AuthorId = pullRequest.AuthorId

	if err := tx.Create(&prInDb).Error; err != nil {
		tx.Rollback()
		return true, ""
	}

	var teamName string
	if err := tx.
		Table("team_members").
		Select("team_name").
		Where("user_id = ?", pullRequest.AuthorId).
		Limit(1).
		Scan(&teamName).Error; err != nil {

		tx.Rollback()
		return true, ""
	}

	var teamUsers []string
	if err := tx.
		Table("team_members").
		Select("user_id").
		Where("team_name = ?", teamName).
		Scan(&teamUsers).Error; err != nil {

		tx.Rollback()
		return true, ""
	}

	var usedUsers = make(map[string]bool)
	for _, userID := range teamUsers {
		var teamUser User
		if err := tx.First(&teamUser, "user_id = ?", userID).Error; err != nil {
			tx.Rollback()
			return true, ""
		}
		if !teamUser.IsActive {
			usedUsers[userID] = true
		}
	}
	usedUsers[prInDb.AuthorId] = true

	for i := 0; i < 2 && len(teamUsers) > len(usedUsers); i++ {
		var newUserIndex = rand.IntN(len(teamUsers) - len(usedUsers))
		var newUserId string
		for _, userId := range teamUsers {
			if usedUsers[userId] {
				continue
			}
			if newUserIndex == 0 {
				newUserId = userId
				break
			}
			newUserIndex--
		}

		tx.Table("pull_request_reviewers").
			Create(map[string]interface{}{
				"pull_request_id": prInDb.PullRequestId,
				"user_id":         newUserId,
			})
	}

	var assignedReviewers []string
	if err := tx.
		Table("pull_request_reviewers").
		Select("user_id").
		Where("pull_request_id = ?", prInDb.PullRequestId).
		Scan(&assignedReviewers).Error; err != nil {

		tx.Rollback()
		return true, ""
	}

	pullRequest.Status = prInDb.Status
	pullRequest.AssignedReviewers = assignedReviewers

	if err := tx.Commit().Error; err != nil {
		return true, ""
	}
	return false, ""
}
