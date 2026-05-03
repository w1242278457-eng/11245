package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/k1LoW/deck"
	"github.com/k1LoW/deck/config"
	"github.com/k1LoW/deck/github"
	"github.com/k1LoW/deck/md"
	"github.com/k1LoW/errors"
	"github.com/spf13/cobra"
)

var (
	githubToken     string
	githubOutput    string
	githubApply     bool
	githubSortField string
	githubSortOrder string
	githubTitle     string
)

var githubCmd = &cobra.Command{
	Use:   "github [owner/repo]",
	Short: "generate GitHub repository report",
	Long: `generate GitHub repository report with issues and pull requests.

The report will be generated in a consistent format with:
- Unified time format (YYYY-MM-DD HH:MM:SS)
- Stable sorting of issues and PRs
- Friendly empty data messages
- Consistent Markdown structure

Examples:
  deck github k1LoW/deck
  deck github k1LoW/deck --output report.md
  deck github k1LoW/deck --apply --title "My Report"
  deck github k1LoW/deck --sort-by created --sort-order desc`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		repoSlug := args[0]

		owner, repo, err := github.ParseRepoSlug(repoSlug)
		if err != nil {
			return err
		}

		client := github.NewClientFromEnv()
		if githubToken != "" {
			client = github.NewClient(githubToken)
		}

		cmd.Printf("Fetching data for %s/%s...\n", owner, repo)

		repository, err := client.GetRepository(ctx, owner, repo)
		if err != nil {
			return fmt.Errorf("failed to fetch repository data: %w", err)
		}

		applySorting(repository)

		cmd.Printf("Issues: %d open, %d closed\n",
			len(repository.OpenIssues()),
			len(repository.ClosedIssues()))
		cmd.Printf("PRs: %d open, %d merged, %d closed\n",
			len(repository.OpenPullRequests()),
			len(repository.MergedPullRequests()),
			len(repository.ClosedPullRequests()))

		generator := github.NewMarkdownGenerator()
		if githubTitle != "" {
			generator.TitleLevel = 1
		}
		report := generator.GenerateRepositoryReport(repository)

		if githubOutput != "" {
			if err := os.WriteFile(githubOutput, []byte(report), 0o644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}
			cmd.Printf("Report written to %s\n", githubOutput)
		} else if !githubApply {
			fmt.Println(report)
		}

		if githubApply {
			return applyReportToSlides(ctx, cmd, report, repo)
		}

		return nil
	},
}

func applySorting(repo *github.Repository) {
	field := github.SortByNumber
	switch strings.ToLower(githubSortField) {
	case "created", "created_at":
		field = github.SortByCreatedAt
	case "updated", "updated_at":
		field = github.SortByUpdatedAt
	case "title":
		field = github.SortByTitle
	case "author":
		field = github.SortByAuthor
	}

	descending := false
	if strings.ToLower(githubSortOrder) == "desc" || strings.ToLower(githubSortOrder) == "descending" {
		descending = true
	}

	switch field {
	case github.SortByCreatedAt:
		github.SortIssuesByCreatedAt(repo.Issues, descending)
		github.SortPRsByCreatedAt(repo.PullRequests, descending)
	case github.SortByUpdatedAt:
		github.SortIssuesByUpdatedAt(repo.Issues, descending)
		github.SortPRsByUpdatedAt(repo.PullRequests, descending)
	case github.SortByTitle:
		github.SortIssuesByNumber(repo.Issues, false)
		github.SortPRsByNumber(repo.PullRequests, false)
	case github.SortByAuthor:
		github.SortIssuesByNumber(repo.Issues, false)
		github.SortPRsByNumber(repo.PullRequests, false)
	default:
		github.SortIssuesByNumber(repo.Issues, descending)
		github.SortPRsByNumber(repo.PullRequests, descending)
	}
}

func applyReportToSlides(ctx context.Context, cmd *cobra.Command, report string, repoName string) error {
	cfg, err := config.Load(profile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var mdFile string
	if githubOutput != "" {
		mdFile = githubOutput
	} else {
		mdFile = filepath.Join(os.TempDir(), fmt.Sprintf("deck-report-%s.md", repoName))
		if err := os.WriteFile(mdFile, []byte(report), 0o644); err != nil {
			return fmt.Errorf("failed to write temporary markdown file: %w", err)
		}
		defer os.Remove(mdFile)
	}

	m, err := md.ParseFile(mdFile, cfg)
	if err != nil {
		return err
	}

	if presentationID == "" && m.Frontmatter != nil && m.Frontmatter.PresentationID != "" {
		presentationID = m.Frontmatter.PresentationID
	}

	if presentationID == "" {
		return fmt.Errorf("presentation ID is required, please specify it with --presentation-id")
	}

	slides, err := m.ToSlides(ctx, codeBlockToImageCmd)
	if err != nil {
		return fmt.Errorf("failed to convert markdown contents to slides: %w", err)
	}

	opts := []deck.Option{
		deck.WithProfile(profile),
		deck.WithPresentationID(presentationID),
	}

	d, err := deck.New(ctx, opts...)
	if err != nil {
		if errors.Is(err, deck.HTTPClientError) {
			cmd.Println(setupInstructionMessage)
		}
		return err
	}

	if githubTitle != "" {
		if err := d.UpdateTitle(ctx, githubTitle); err != nil {
			return err
		}
	}

	if err := d.Apply(ctx, slides); err != nil {
		return err
	}

	cmd.Println(color.GreenString("Report applied to presentation successfully!"))
	return nil
}

func init() {
	rootCmd.AddCommand(githubCmd)
	githubCmd.Flags().StringVarP(&githubToken, "token", "", "", "GitHub token (can also be set via GITHUB_TOKEN environment variable)")
	githubCmd.Flags().StringVarP(&githubOutput, "output", "o", "", "output markdown file (default: stdout)")
	githubCmd.Flags().BoolVarP(&githubApply, "apply", "a", false, "apply report to Google Slides presentation")
	githubCmd.Flags().StringVarP(&githubSortField, "sort-by", "", "number", "sort field: number, created, updated, title, author")
	githubCmd.Flags().StringVarP(&githubSortOrder, "sort-order", "", "asc", "sort order: asc, ascending, desc, descending")
	githubCmd.Flags().StringVarP(&presentationID, "presentation-id", "i", "", "Google Slides presentation ID (for --apply)")
	githubCmd.Flags().StringVarP(&githubTitle, "title", "t", "", "title of the presentation (for --apply)")
	githubCmd.Flags().StringVarP(&codeBlockToImageCmd, "code-block-to-image-command", "c", "", "command to convert code blocks to images")
}
