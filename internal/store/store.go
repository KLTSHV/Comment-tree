package store

import "github.com/KLTSHV/Comment-tree/internal/domain"

type Repository interface {
	Create(parentID int64, author, text string) (*domain.Comment, error)
	GetComment(id int64) (*domain.Comment, bool)
	DeleteSubtree(id int64) (deletedCount int, err error)

	// Для дерева
	ListChildIDs(parentID int64) []int64
	// Для поиска
	SnapshotAll() []domain.Comment
}
