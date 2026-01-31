package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/KLTSHV/Comment-tree/internal/domain"
	"github.com/KLTSHV/Comment-tree/internal/search"
	"github.com/KLTSHV/Comment-tree/internal/store"
	"github.com/KLTSHV/Comment-tree/internal/tree"
)

/*
HTTP API слой:
- POST /comments
- GET  /comments?parent={id}&page=&limit=&sort=
- DELETE /comments/{id}
- GET /comments/search?q=&page=&limit=&sort=
*/

func RegisterRoutes(mux *http.ServeMux, repo store.Repository) {
	mux.HandleFunc("/comments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleCreate(repo, w, r)
		case http.MethodGet:
			handleGetTree(repo, w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/comments/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handleSearch(repo, w, r)
	})

	mux.HandleFunc("/comments/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handleDelete(repo, w, r)
	})
}

func WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleCreate(repo store.Repository, w http.ResponseWriter, r *http.Request) {
	var req domain.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	c, err := repo.Create(req.ParentID, req.Author, req.Text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusCreated, c)
}

func handleGetTree(repo store.Repository, w http.ResponseWriter, r *http.Request) {
	parentStr := strings.TrimSpace(r.URL.Query().Get("parent"))

	sortMode := strings.TrimSpace(r.URL.Query().Get("sort"))
	if sortMode == "" {
		sortMode = "created_at_asc"
	}

	page := parseIntQuery(r, "page", 1)
	limit := parseIntQuery(r, "limit", 20)

	// parent отсутствует или 0 => корневые, каждый корень с полным поддеревом
	if parentStr == "" || parentStr == "0" {
		rootIDs := repo.ListChildIDs(0)

		// сортировка корней через tree builder helper
		// сортируем здесь через временный node-build: сортируем ID по created_at.
		rootIDs = sortIDsByCreatedAt(repo, rootIDs, sortMode)

		total := len(rootIDs)
		pagedRoots := paginateIDs(rootIDs, page, limit)

		items := make([]domain.CommentNode, 0, len(pagedRoots))
		for _, id := range pagedRoots {
			node, err := tree.BuildNode(repo, id, sortMode)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			items = append(items, node)
		}

		resp := domain.TreeResponse{
			Items: items,
			Meta:  domain.PageMeta{Page: normPage(page), Limit: normLimit(limit), Total: total},
			Sort:  sortMode,
		}
		writeJSON(w, http.StatusOK, resp)
		return
	}

	// parent задан => вернуть parent-узел. его прямые дети постранично, вглубь полностью
	parentID, err := strconv.ParseInt(parentStr, 10, 64)
	if err != nil || parentID < 0 {
		http.Error(w, "invalid parent id", http.StatusBadRequest)
		return
	}

	node, meta, err := tree.BuildNodeWithPagedChildren(repo, parentID, sortMode, page, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp := domain.TreeResponse{
		Parent: &node,
		Items:  nil,
		Meta:   meta,
		Sort:   sortMode,
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleDelete(repo store.Repository, w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/comments/"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	n, err := repo.DeleteSubtree(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":      id,
		"deleted": n,
	})
}

func handleSearch(repo store.Repository, w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))

	sortMode := strings.TrimSpace(r.URL.Query().Get("sort"))
	if sortMode == "" {
		sortMode = "created_at_desc"
	}

	page := parseIntQuery(r, "page", 1)
	limit := parseIntQuery(r, "limit", 20)

	items, meta := search.Find(repo, q, sortMode, page, limit)

	resp := domain.SearchResponse{
		Query: q,
		Items: items,
		Meta:  meta,
		Sort:  sortMode,
	}
	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func parseIntQuery(r *http.Request, key string, def int) int {
	v := strings.TrimSpace(r.URL.Query().Get(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
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

// сортировка ID по created_at
func sortIDsByCreatedAt(repo store.Repository, ids []int64, sortMode string) []int64 {
	out := make([]int64, len(ids))
	copy(out, ids)

	desc := strings.EqualFold(sortMode, "created_at_desc")

	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			ci, _ := repo.GetComment(out[i])
			cj, _ := repo.GetComment(out[j])
			if ci == nil || cj == nil {
				continue
			}
			needSwap := false
			if !desc && cj.CreatedAt.Before(ci.CreatedAt) {
				needSwap = true
			}
			if desc && cj.CreatedAt.After(ci.CreatedAt) {
				needSwap = true
			}
			if needSwap {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}
