package entity

import "time"

type RoleApplicationStatus string

const (
	RoleApplicationPending  RoleApplicationStatus = "pending"
	RoleApplicationApproved RoleApplicationStatus = "approved"
	RoleApplicationRejected RoleApplicationStatus = "rejected"
)

type RoleApplication struct {
	ID            string
	UserID        string
	RequestedRole Role
	Comment       string
	Status        RoleApplicationStatus
	ReviewerID    string
	CreatedAt     time.Time
	ReviewedAt    time.Time
}

func (a *RoleApplication) Approve(reviewerID string) {
	a.Status = RoleApplicationApproved
	a.ReviewerID = reviewerID
	a.ReviewedAt = time.Now().UTC()
}

func (a *RoleApplication) Reject(reviewerID string) {
	a.Status = RoleApplicationRejected
	a.ReviewerID = reviewerID
	a.ReviewedAt = time.Now().UTC()
}

func (a *RoleApplication) IsPending() bool {
	return a.Status == RoleApplicationPending
}