package github

import (
	"cmp"
	"slices"
)

type SortField int

const (
	SortByNumber SortField = iota
	SortByCreatedAt
	SortByUpdatedAt
	SortByTitle
	SortByAuthor
)

type SortOrder int

const (
	SortAscending SortOrder = iota
	SortDescending
)

type IssueSorter struct {
	Field SortField
	Order SortOrder
}

func (s *IssueSorter) Sort(issues []*Issue) {
	if len(issues) <= 1 {
		return
	}

	slices.SortStableFunc(issues, func(a, b *Issue) int {
		var result int

		switch s.Field {
		case SortByNumber:
			result = cmp.Compare(a.Number, b.Number)
		case SortByCreatedAt:
			result = a.CreatedAt.Compare(b.CreatedAt)
		case SortByUpdatedAt:
			result = a.UpdatedAt.Compare(b.UpdatedAt)
		case SortByTitle:
			result = cmp.Compare(a.Title, b.Title)
		case SortByAuthor:
			result = cmp.Compare(a.Author, b.Author)
		default:
			result = cmp.Compare(a.Number, b.Number)
		}

		if s.Order == SortDescending {
			result = -result
		}

		if result == 0 {
			result = cmp.Compare(a.Number, b.Number)
		}

		return result
	})
}

func SortIssuesByNumber(issues []*Issue, descending bool) {
	sorter := &IssueSorter{
		Field: SortByNumber,
		Order: SortAscending,
	}
	if descending {
		sorter.Order = SortDescending
	}
	sorter.Sort(issues)
}

func SortIssuesByCreatedAt(issues []*Issue, descending bool) {
	sorter := &IssueSorter{
		Field: SortByCreatedAt,
		Order: SortAscending,
	}
	if descending {
		sorter.Order = SortDescending
	}
	sorter.Sort(issues)
}

func SortIssuesByUpdatedAt(issues []*Issue, descending bool) {
	sorter := &IssueSorter{
		Field: SortByUpdatedAt,
		Order: SortAscending,
	}
	if descending {
		sorter.Order = SortDescending
	}
	sorter.Sort(issues)
}

type PRSorter struct {
	Field SortField
	Order SortOrder
}

func (s *PRSorter) Sort(prs []*PullRequest) {
	if len(prs) <= 1 {
		return
	}

	slices.SortStableFunc(prs, func(a, b *PullRequest) int {
		var result int

		switch s.Field {
		case SortByNumber:
			result = cmp.Compare(a.Number, b.Number)
		case SortByCreatedAt:
			result = a.CreatedAt.Compare(b.CreatedAt)
		case SortByUpdatedAt:
			result = a.UpdatedAt.Compare(b.UpdatedAt)
		case SortByTitle:
			result = cmp.Compare(a.Title, b.Title)
		case SortByAuthor:
			result = cmp.Compare(a.Author, b.Author)
		default:
			result = cmp.Compare(a.Number, b.Number)
		}

		if s.Order == SortDescending {
			result = -result
		}

		if result == 0 {
			result = cmp.Compare(a.Number, b.Number)
		}

		return result
	})
}

func SortPRsByNumber(prs []*PullRequest, descending bool) {
	sorter := &PRSorter{
		Field: SortByNumber,
		Order: SortAscending,
	}
	if descending {
		sorter.Order = SortDescending
	}
	sorter.Sort(prs)
}

func SortPRsByCreatedAt(prs []*PullRequest, descending bool) {
	sorter := &PRSorter{
		Field: SortByCreatedAt,
		Order: SortAscending,
	}
	if descending {
		sorter.Order = SortDescending
	}
	sorter.Sort(prs)
}

func SortPRsByUpdatedAt(prs []*PullRequest, descending bool) {
	sorter := &PRSorter{
		Field: SortByUpdatedAt,
		Order: SortAscending,
	}
	if descending {
		sorter.Order = SortDescending
	}
	sorter.Sort(prs)
}

func SortPRsByStatusAndNumber(prs []*PullRequest) {
	if len(prs) <= 1 {
		return
	}

	slices.SortStableFunc(prs, func(a, b *PullRequest) int {
		statusOrder := map[string]int{
			"open":   0,
			"draft":  1,
			"merged": 2,
			"closed": 3,
		}

		aStatus := a.Status()
		bStatus := b.Status()

		aOrder := statusOrder[aStatus]
		bOrder := statusOrder[bStatus]

		if aOrder != bOrder {
			return cmp.Compare(aOrder, bOrder)
		}

		return cmp.Compare(a.Number, b.Number)
	})
}
