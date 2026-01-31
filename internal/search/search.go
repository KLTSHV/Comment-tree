package search

import (
	"sort"
	"strings"

	"github.com/KLTSHV/Comment-tree/internal/domain"
	"github.com/KLTSHV/Comment-tree/internal/store"
)

func Find(repo store.Repository, q string, sortMode string, page, limit int) (items []domain.Comment, meta domain.PageMeta) {
	q = strings.TrimSpace(q)
	page = normPage(page)
	limit = normLimit(limit)

	if q == "" {
		return []domain.Comment{}, domain.PageMeta{Page: page, Limit: limit, Total: 0}
	}

	qLower := strings.ToLower(q)
	all := repo.SnapshotAll()

	matches := make([]domain.Comment, 0)
	for _, c := range all {
		if strings.Contains(strings.ToLower(c.Text), qLower) ||
			strings.Contains(strings.ToLower(c.Author), qLower) {
			matches = append(matches, c)
		}
	}

	asc := true
	if strings.EqualFold(sortMode, "created_at_desc") {
		asc = false
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if asc {
			return matches[i].CreatedAt.Before(matches[j].CreatedAt)
		}
		return matches[i].CreatedAt.After(matches[j].CreatedAt)
	})

	total := len(matches)
	out := paginate(matches, page, limit)

	return out, domain.PageMeta{Page: page, Limit: limit, Total: total}
}

func paginate[T any](arr []T, page, limit int) []T {
	start := (page - 1) * limit
	if start >= len(arr) {
		return []T{}
	}
	end := start + limit
	if end > len(arr) {
		end = len(arr)
	}
	return arr[start:end]
}

func normLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	return limit
}

func normPage(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}
