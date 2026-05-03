package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	GitHubAPIURL = "https://api.github.com"
	PerPage      = 100
)

type Client struct {
	token      string
	httpClient *http.Client
	baseURL    string
}

func NewClient(token string) *Client {
	return &Client{
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    GitHubAPIURL,
	}
}

func NewClientFromEnv() *Client {
	token := os.Getenv("GITHUB_TOKEN")
	return NewClient(token)
}

func (c *Client) newRequest(ctx context.Context, method, path string) (*http.Request, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "deck/1.0")

	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	return req, nil
}

func (c *Client) do(req *http.Request, v any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(body))
	}

	if v != nil {
		return json.NewDecoder(resp.Body).Decode(v)
	}

	return nil
}

type githubIssue struct {
	Number    int           `json:"number"`
	Title     string        `json:"title"`
	State     string        `json:"state"`
	User      githubUser    `json:"user"`
	CreatedAt string        `json:"created_at"`
	UpdatedAt string        `json:"updated_at"`
	ClosedAt  *string       `json:"closed_at"`
	HTMLURL   string        `json:"html_url"`
	Labels    []githubLabel `json:"labels"`
	Body      string        `json:"body"`
	PullRequest *any        `json:"pull_request"`
}

type githubPullRequest struct {
	Number      int           `json:"number"`
	Title       string        `json:"title"`
	State       string        `json:"state"`
	User        githubUser    `json:"user"`
	CreatedAt   string        `json:"created_at"`
	UpdatedAt   string        `json:"updated_at"`
	ClosedAt    *string       `json:"closed_at"`
	MergedAt    *string       `json:"merged_at"`
	HTMLURL     string        `json:"html_url"`
	Labels      []githubLabel `json:"labels"`
	Body        string        `json:"body"`
	Draft       bool          `json:"draft"`
	Merged      bool          `json:"merged"`
	Additions   int           `json:"additions"`
	Deletions   int           `json:"deletions"`
	ChangedFiles int          `json:"changed_files"`
}

type githubUser struct {
	Login string `json:"login"`
}

type githubLabel struct {
	Name string `json:"name"`
}

func parseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}

	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}

	if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Unix(ts, 0), nil
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}

func (c *Client) GetIssues(ctx context.Context, owner, repo string) ([]*Issue, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues?state=all&per_page=%d", owner, repo, PerPage)
	req, err := c.newRequest(ctx, "GET", path)
	if err != nil {
		return nil, err
	}

	var ghIssues []githubIssue
	if err := c.do(req, &ghIssues); err != nil {
		return nil, err
	}

	var issues []*Issue
	for _, ghi := range ghIssues {
		if ghi.PullRequest != nil {
			continue
		}

		issue := &Issue{
			Number:  ghi.Number,
			Title:   ghi.Title,
			State:   ghi.State,
			Author:  ghi.User.Login,
			HTMLURL: ghi.HTMLURL,
			Body:    ghi.Body,
		}

		if t, err := parseTime(ghi.CreatedAt); err == nil {
			issue.CreatedAt = t
		}
		if t, err := parseTime(ghi.UpdatedAt); err == nil {
			issue.UpdatedAt = t
		}
		if ghi.ClosedAt != nil {
			if t, err := parseTime(*ghi.ClosedAt); err == nil {
				issue.ClosedAt = &t
			}
		}

		for _, label := range ghi.Labels {
			issue.Labels = append(issue.Labels, label.Name)
		}

		issues = append(issues, issue)
	}

	SortIssuesByNumber(issues, false)
	return issues, nil
}

func (c *Client) GetPullRequests(ctx context.Context, owner, repo string) ([]*PullRequest, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls?state=all&per_page=%d", owner, repo, PerPage)
	req, err := c.newRequest(ctx, "GET", path)
	if err != nil {
		return nil, err
	}

	var ghPRs []githubPullRequest
	if err := c.do(req, &ghPRs); err != nil {
		return nil, err
	}

	var prs []*PullRequest
	for _, ghpr := range ghPRs {
		pr := &PullRequest{
			Number:       ghpr.Number,
			Title:        ghpr.Title,
			State:        ghpr.State,
			Author:       ghpr.User.Login,
			IsDraft:      ghpr.Draft,
			IsMerged:     ghpr.Merged,
			HTMLURL:      ghpr.HTMLURL,
			Body:         ghpr.Body,
			Additions:    ghpr.Additions,
			Deletions:    ghpr.Deletions,
			ChangedFiles: ghpr.ChangedFiles,
		}

		if t, err := parseTime(ghpr.CreatedAt); err == nil {
			pr.CreatedAt = t
		}
		if t, err := parseTime(ghpr.UpdatedAt); err == nil {
			pr.UpdatedAt = t
		}
		if ghpr.ClosedAt != nil {
			if t, err := parseTime(*ghpr.ClosedAt); err == nil {
				pr.ClosedAt = &t
			}
		}
		if ghpr.MergedAt != nil {
			if t, err := parseTime(*ghpr.MergedAt); err == nil {
				pr.MergedAt = &t
			}
		}

		for _, label := range ghpr.Labels {
			pr.Labels = append(pr.Labels, label.Name)
		}

		prs = append(prs, pr)
	}

	SortPRsByNumber(prs, false)
	return prs, nil
}

func (c *Client) GetRepository(ctx context.Context, owner, repo string) (*Repository, error) {
	issues, err := c.GetIssues(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get issues: %w", err)
	}

	prs, err := c.GetPullRequests(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull requests: %w", err)
	}

	return &Repository{
		Owner:        owner,
		Name:         repo,
		FullName:     fmt.Sprintf("%s/%s", owner, repo),
		HTMLURL:      fmt.Sprintf("https://github.com/%s/%s", owner, repo),
		Issues:       issues,
		PullRequests: prs,
	}, nil
}

func ParseRepoSlug(slug string) (owner, repo string, err error) {
	slug = strings.TrimSpace(slug)
	slug = strings.TrimPrefix(slug, "https://github.com/")
	slug = strings.TrimPrefix(slug, "github.com/")
	slug = strings.TrimSuffix(slug, ".git")
	slug = strings.TrimSuffix(slug, "/")

	parts := strings.Split(slug, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repository slug: %s (expected format: owner/repo)", slug)
	}

	return parts[0], parts[1], nil
}
