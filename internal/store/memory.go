package store

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/KLTSHV/Comment-tree/internal/domain"
)

type MemoryStore struct {
	mu       sync.RWMutex
	nextID   int64
	byID     map[int64]*domain.Comment
	children map[int64][]int64 // parentID -> childID
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		nextID:   1,
		byID:     make(map[int64]*domain.Comment),
		children: make(map[int64][]int64),
	}
}

func (s *MemoryStore) Create(parentID int64, author, text string) (*domain.Comment, error) {
	author = strings.TrimSpace(author)
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("text is required")
	}
	if author == "" {
		author = "anonymous"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if parentID != 0 {
		if _, ok := s.byID[parentID]; !ok {
			return nil, fmt.Errorf("parent comment %d not found", parentID)
		}
	}

	id := s.nextID
	s.nextID++

	c := &domain.Comment{
		ID:        id,
		ParentID:  parentID,
		Author:    author,
		Text:      text,
		CreatedAt: time.Now().UTC(),
	}

	s.byID[id] = c
	s.children[parentID] = append(s.children[parentID], id)

	cp := *c
	return &cp, nil
}

func (s *MemoryStore) GetComment(id int64) (*domain.Comment, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, ok := s.byID[id]
	if !ok {
		return nil, false
	}
	cp := *c
	return &cp, true
}

func (s *MemoryStore) ListChildIDs(parentID int64) []int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := s.children[parentID]
	out := make([]int64, len(ids))
	copy(out, ids)
	return out
}

func (s *MemoryStore) SnapshotAll() []domain.Comment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]domain.Comment, 0, len(s.byID))
	for _, c := range s.byID {
		out = append(out, *c)
	}
	return out
}

func (s *MemoryStore) DeleteSubtree(id int64) (deletedCount int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.byID[id]
	if !ok {
		return 0, fmt.Errorf("comment %d not found", id)
	}

	// отвязать от родителя
	parentID := c.ParentID
	s.children[parentID] = filterOut(s.children[parentID], id)

	// DFS по поддереву
	stack := []int64{id}

	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		for _, ch := range s.children[cur] {
			stack = append(stack, ch)
		}
		delete(s.children, cur)
		delete(s.byID, cur)
		deletedCount++
	}

	return deletedCount, nil
}

func filterOut(arr []int64, x int64) []int64 {
	out := arr[:0]
	for _, v := range arr {
		if v != x {
			out = append(out, v)
		}
	}
	return out
}
