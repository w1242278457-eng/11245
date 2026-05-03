package github

import (
	"fmt"
	"time"
)

const (
	TimeFormat = "2006-01-02 15:04:05"
)

func FormatTime(t time.Time) string {
	return t.Format(TimeFormat)
}

func FormatTimeFromUnix(ts int64) string {
	return time.Unix(ts, 0).Format(TimeFormat)
}

func ParseRFC3339(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

func FormatTimeRFC3339(t time.Time) string {
	return t.Format(time.RFC3339)
}

type Issue struct {
	Number    int
	Title     string
	State     string
	Author    string
	CreatedAt time.Time
	UpdatedAt time.Time
	ClosedAt  *time.Time
	HTMLURL   string
	Labels    []string
	Body      string
}

func (i *Issue) FormattedCreatedAt() string {
	return FormatTime(i.CreatedAt)
}

func (i *Issue) FormattedUpdatedAt() string {
	return FormatTime(i.UpdatedAt)
}

func (i *Issue) FormattedClosedAt() string {
	if i.ClosedAt == nil {
		return ""
	}
	return FormatTime(*i.ClosedAt)
}

type PullRequest struct {
	Number      int
	Title       string
	State       string
	Author      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ClosedAt    *time.Time
	MergedAt    *time.Time
	HTMLURL     string
	Labels      []string
	Body        string
	IsDraft     bool
	IsMerged    bool
	Additions   int
	Deletions   int
	ChangedFiles int
}

func (pr *PullRequest) FormattedCreatedAt() string {
	return FormatTime(pr.CreatedAt)
}

func (pr *PullRequest) FormattedUpdatedAt() string {
	return FormatTime(pr.UpdatedAt)
}

func (pr *PullRequest) FormattedClosedAt() string {
	if pr.ClosedAt == nil {
		return ""
	}
	return FormatTime(*pr.ClosedAt)
}

func (pr *PullRequest) FormattedMergedAt() string {
	if pr.MergedAt == nil {
		return ""
	}
	return FormatTime(*pr.MergedAt)
}

func (pr *PullRequest) Status() string {
	if pr.IsMerged {
		return "merged"
	}
	if pr.IsDraft {
		return "draft"
	}
	return pr.State
}

func (pr *PullRequest) DiffStats() string {
	return fmt.Sprintf("+%d -%d (%d files)", pr.Additions, pr.Deletions, pr.ChangedFiles)
}

type Repository struct {
	Owner      string
	Name       string
	FullName   string
	HTMLURL    string
	Issues     []*Issue
	PullRequests []*PullRequest
}

func (r *Repository) HasIssues() bool {
	return len(r.Issues) > 0
}

func (r *Repository) HasPullRequests() bool {
	return len(r.PullRequests) > 0
}

func (r *Repository) OpenIssues() []*Issue {
	var open []*Issue
	for _, issue := range r.Issues {
		if issue.State == "open" {
			open = append(open, issue)
		}
	}
	return open
}

func (r *Repository) ClosedIssues() []*Issue {
	var closed []*Issue
	for _, issue := range r.Issues {
		if issue.State == "closed" {
			closed = append(closed, issue)
		}
	}
	return closed
}

func (r *Repository) OpenPullRequests() []*PullRequest {
	var open []*PullRequest
	for _, pr := range r.PullRequests {
		if pr.State == "open" && !pr.IsMerged {
			open = append(open, pr)
		}
	}
	return open
}

func (r *Repository) MergedPullRequests() []*PullRequest {
	var merged []*PullRequest
	for _, pr := range r.PullRequests {
		if pr.IsMerged {
			merged = append(merged, pr)
		}
	}
	return merged
}

func (r *Repository) ClosedPullRequests() []*PullRequest {
	var closed []*PullRequest
	for _, pr := range r.PullRequests {
		if pr.State == "closed" && !pr.IsMerged {
			closed = append(closed, pr)
		}
	}
	return closed
}
