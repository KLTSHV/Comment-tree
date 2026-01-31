package tree

import (
	"fmt"
	"sort"
	"strings"

	"github.com/KLTSHV/Comment-tree/internal/domain"
	"github.com/KLTSHV/Comment-tree/internal/store"
)

// BuildNode строит дерево комментариев с корнем id
func BuildNode(repo store.Repository, id int64, sortMode string) (domain.CommentNode, error) {
	c, ok := repo.GetComment(id)
	if !ok {
		return domain.CommentNode{}, fmt.Errorf("comment %d not found", id)
	}

	node := domain.CommentNode{Comment: *c}

	childIDs := repo.ListChildIDs(id)
	sortIDsByCreatedAt(repo, childIDs, sortMode)

	for _, chID := range childIDs {
		chNode, err := BuildNode(repo, chID, sortMode)
		if err != nil {
			return domain.CommentNode{}, err
		}
		node.Children = append(node.Children, chNode)
	}

	return node, nil
}

// BuildNodeWithPagedChildren у parent прямые дети пагинируются но каждое дитя раскрывается полностью вниз
func BuildNodeWithPagedChildren(repo store.Repository, id int64, sortMode string, page, limit int) (domain.CommentNode, domain.PageMeta, error) {
	c, ok := repo.GetComment(id)
	if !ok {
		return domain.CommentNode{}, domain.PageMeta{}, fmt.Errorf("comment %d not found", id)
	}

	node := domain.CommentNode{Comment: *c}

	childIDs := repo.ListChildIDs(id)
	sortIDsByCreatedAt(repo, childIDs, sortMode)

	total := len(childIDs)
	paged := paginateIDs(childIDs, page, limit)

	meta := domain.PageMeta{
		Page:  normPage(page),
		Limit: normLimit(limit),
		Total: total,
	}

	for _, chID := range paged {
		chNode, err := BuildNode(repo, chID, sortMode)
		if err != nil {
			return domain.CommentNode{}, domain.PageMeta{}, err
		}
		node.Children = append(node.Children, chNode)
	}

	return node, meta, nil
}

func sortIDsByCreatedAt(repo store.Repository, ids []int64, sortMode string) {
	asc := true
	if strings.EqualFold(sortMode, "created_at_desc") {
		asc = false
	}

	sort.SliceStable(ids, func(i, j int) bool {
		ci, _ := repo.GetComment(ids[i])
		cj, _ := repo.GetComment(ids[j])
		if ci == nil || cj == nil {
			return asc
		}
		if asc {
			return ci.CreatedAt.Before(cj.CreatedAt)
		}
		return ci.CreatedAt.After(cj.CreatedAt)
	})
}

func paginateIDs(ids []int64, page, limit int) []int64 {
	limit = normLimit(limit)
	page = normPage(page)

	start := (page - 1) * limit
	if start >= len(ids) {
		return []int64{}
	}
	end := start + limit
	if end > len(ids) {
		end = len(ids)
	}
	return ids[start:end]
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
