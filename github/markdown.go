package github

import (
	"fmt"
	"strings"
	"time"
)

const (
	NoIssuesMessage    = "暂无 Issues"
	NoPRsMessage       = "暂无 Pull Requests"
	NoOpenIssuesMessage = "暂无打开的 Issues"
	NoOpenPRsMessage    = "暂无打开的 Pull Requests"
	NoMergedPRsMessage  = "暂无已合并的 Pull Requests"
	NoClosedPRsMessage  = "暂无已关闭的 Pull Requests"
)

type EmptyDataHandler struct {
	ShowMessage bool
	CustomMessage map[string]string
}

func NewEmptyDataHandler() *EmptyDataHandler {
	return &EmptyDataHandler{
		ShowMessage: true,
		CustomMessage: make(map[string]string),
	}
}

func (h *EmptyDataHandler) GetMessage(key string, defaultMsg string) string {
	if msg, ok := h.CustomMessage[key]; ok {
		return msg
	}
	return defaultMsg
}

func (h *EmptyDataHandler) FormatEmptyList(items any, emptyMsg string) string {
	switch v := items.(type) {
	case []*Issue:
		if len(v) == 0 {
			if h.ShowMessage {
				return h.GetMessage("issues", emptyMsg)
			}
			return ""
		}
	case []*PullRequest:
		if len(v) == 0 {
			if h.ShowMessage {
				return h.GetMessage("prs", emptyMsg)
			}
			return ""
		}
	}
	return ""
}

type MarkdownGenerator struct {
	TitleLevel     int
	IncludeDetails bool
	TimeFormat     string
	EmptyHandler   *EmptyDataHandler
}

func NewMarkdownGenerator() *MarkdownGenerator {
	return &MarkdownGenerator{
		TitleLevel:     1,
		IncludeDetails: true,
		TimeFormat:     TimeFormat,
		EmptyHandler:   NewEmptyDataHandler(),
	}
}

func (g *MarkdownGenerator) header(level int, text string) string {
	return fmt.Sprintf("%s %s\n", strings.Repeat("#", level), text)
}

func (g *MarkdownGenerator) formatTime(t time.Time) string {
	if g.TimeFormat == "" {
		return FormatTime(t)
	}
	return t.Format(g.TimeFormat)
}

func (g *MarkdownGenerator) GenerateIssue(issue *Issue, level int) string {
	if issue == nil {
		return ""
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s [#%d] %s\n",
		g.header(level, ""),
		issue.Number,
		issue.Title))

	sb.WriteString(fmt.Sprintf("- **状态**: %s\n", issue.State))
	sb.WriteString(fmt.Sprintf("- **作者**: %s\n", issue.Author))
	sb.WriteString(fmt.Sprintf("- **创建时间**: %s\n", g.formatTime(issue.CreatedAt)))
	sb.WriteString(fmt.Sprintf("- **更新时间**: %s\n", g.formatTime(issue.UpdatedAt)))

	if issue.ClosedAt != nil {
		sb.WriteString(fmt.Sprintf("- **关闭时间**: %s\n", g.formatTime(*issue.ClosedAt)))
	}

	if len(issue.Labels) > 0 {
		sb.WriteString(fmt.Sprintf("- **标签**: %s\n", strings.Join(issue.Labels, ", ")))
	}

	sb.WriteString(fmt.Sprintf("- **链接**: [%s](%s)\n", issue.HTMLURL, issue.HTMLURL))

	if g.IncludeDetails && issue.Body != "" {
		sb.WriteString("\n**描述**:\n\n")
		sb.WriteString(issue.Body)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	return sb.String()
}

func (g *MarkdownGenerator) GeneratePullRequest(pr *PullRequest, level int) string {
	if pr == nil {
		return ""
	}

	var sb strings.Builder

	status := pr.Status()
	statusIcon := ""
	switch status {
	case "open":
		statusIcon = "🟢"
	case "draft":
		statusIcon = "⚪"
	case "merged":
		statusIcon = "🟣"
	case "closed":
		statusIcon = "🔴"
	}

	sb.WriteString(fmt.Sprintf("%s [#%d] %s %s\n",
		g.header(level, ""),
		pr.Number,
		statusIcon,
		pr.Title))

	sb.WriteString(fmt.Sprintf("- **状态**: %s\n", status))
	sb.WriteString(fmt.Sprintf("- **作者**: %s\n", pr.Author))
	sb.WriteString(fmt.Sprintf("- **创建时间**: %s\n", g.formatTime(pr.CreatedAt)))
	sb.WriteString(fmt.Sprintf("- **更新时间**: %s\n", g.formatTime(pr.UpdatedAt)))

	if pr.MergedAt != nil {
		sb.WriteString(fmt.Sprintf("- **合并时间**: %s\n", g.formatTime(*pr.MergedAt)))
	} else if pr.ClosedAt != nil {
		sb.WriteString(fmt.Sprintf("- **关闭时间**: %s\n", g.formatTime(*pr.ClosedAt)))
	}

	if len(pr.Labels) > 0 {
		sb.WriteString(fmt.Sprintf("- **标签**: %s\n", strings.Join(pr.Labels, ", ")))
	}

	sb.WriteString(fmt.Sprintf("- **变更**: %s\n", pr.DiffStats()))
	sb.WriteString(fmt.Sprintf("- **链接**: [%s](%s)\n", pr.HTMLURL, pr.HTMLURL))

	if g.IncludeDetails && pr.Body != "" {
		sb.WriteString("\n**描述**:\n\n")
		sb.WriteString(pr.Body)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	return sb.String()
}

func (g *MarkdownGenerator) GenerateIssueList(issues []*Issue, title string, level int) string {
	var sb strings.Builder

	if title != "" {
		sb.WriteString(g.header(level, title))
		sb.WriteString("\n")
	}

	emptyMsg := g.EmptyHandler.FormatEmptyList(issues, NoIssuesMessage)
	if emptyMsg != "" {
		sb.WriteString(fmt.Sprintf("> %s\n\n", emptyMsg))
		return sb.String()
	}

	for _, issue := range issues {
		sb.WriteString(g.GenerateIssue(issue, level+1))
	}

	return sb.String()
}

func (g *MarkdownGenerator) GeneratePullRequestList(prs []*PullRequest, title string, level int) string {
	var sb strings.Builder

	if title != "" {
		sb.WriteString(g.header(level, title))
		sb.WriteString("\n")
	}

	emptyMsg := g.EmptyHandler.FormatEmptyList(prs, NoPRsMessage)
	if emptyMsg != "" {
		sb.WriteString(fmt.Sprintf("> %s\n\n", emptyMsg))
		return sb.String()
	}

	for _, pr := range prs {
		sb.WriteString(g.GeneratePullRequest(pr, level+1))
	}

	return sb.String()
}

func (g *MarkdownGenerator) GenerateRepositoryReport(repo *Repository) string {
	var sb strings.Builder

	sb.WriteString(g.header(g.TitleLevel, fmt.Sprintf("%s 报告", repo.FullName)))
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("- **仓库**: [%s](%s)\n", repo.FullName, repo.HTMLURL))
	sb.WriteString(fmt.Sprintf("- **Issues 总数**: %d\n", len(repo.Issues)))
	sb.WriteString(fmt.Sprintf("- **PR 总数**: %d\n", len(repo.PullRequests)))
	sb.WriteString("\n")

	openIssues := repo.OpenIssues()
	closedIssues := repo.ClosedIssues()
	openPRs := repo.OpenPullRequests()
	mergedPRs := repo.MergedPullRequests()
	closedPRs := repo.ClosedPullRequests()

	sb.WriteString(g.header(g.TitleLevel+1, "Issues 统计"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("- 打开: %d\n", len(openIssues)))
	sb.WriteString(fmt.Sprintf("- 关闭: %d\n", len(closedIssues)))
	sb.WriteString("\n")

	if len(openIssues) > 0 {
		SortIssuesByNumber(openIssues, false)
		sb.WriteString(g.GenerateIssueList(openIssues, "打开的 Issues", g.TitleLevel+1))
	} else {
		sb.WriteString(g.header(g.TitleLevel+1, "打开的 Issues"))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("> %s\n\n", g.EmptyHandler.GetMessage("open_issues", NoOpenIssuesMessage)))
	}

	if len(closedIssues) > 0 {
		SortIssuesByUpdatedAt(closedIssues, true)
		sb.WriteString(g.GenerateIssueList(closedIssues, "关闭的 Issues", g.TitleLevel+1))
	}

	sb.WriteString(g.header(g.TitleLevel+1, "Pull Requests 统计"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("- 打开: %d\n", len(openPRs)))
	sb.WriteString(fmt.Sprintf("- 已合并: %d\n", len(mergedPRs)))
	sb.WriteString(fmt.Sprintf("- 已关闭: %d\n", len(closedPRs)))
	sb.WriteString("\n")

	SortPRsByStatusAndNumber(prsToList(openPRs, mergedPRs, closedPRs))

	if len(openPRs) > 0 {
		SortPRsByNumber(openPRs, false)
		sb.WriteString(g.GeneratePullRequestList(openPRs, "打开的 Pull Requests", g.TitleLevel+1))
	} else {
		sb.WriteString(g.header(g.TitleLevel+1, "打开的 Pull Requests"))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("> %s\n\n", g.EmptyHandler.GetMessage("open_prs", NoOpenPRsMessage)))
	}

	if len(mergedPRs) > 0 {
		SortPRsByUpdatedAt(mergedPRs, true)
		sb.WriteString(g.GeneratePullRequestList(mergedPRs, "已合并的 Pull Requests", g.TitleLevel+1))
	} else {
		sb.WriteString(g.header(g.TitleLevel+1, "已合并的 Pull Requests"))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("> %s\n\n", g.EmptyHandler.GetMessage("merged_prs", NoMergedPRsMessage)))
	}

	if len(closedPRs) > 0 {
		SortPRsByUpdatedAt(closedPRs, true)
		sb.WriteString(g.GeneratePullRequestList(closedPRs, "已关闭的 Pull Requests", g.TitleLevel+1))
	}

	return sb.String()
}

func prsToList(lists ...[]*PullRequest) []*PullRequest {
	var result []*PullRequest
	for _, list := range lists {
		result = append(result, list...)
	}
	return result
}
