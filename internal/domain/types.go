package domain

import "time"

type Comment struct {
	ID        int64     `json:"id"`
	ParentID  int64     `json:"parent_id"`
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type CommentNode struct {
	Comment
	Children []CommentNode `json:"children"`
}

type PageMeta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type TreeResponse struct {
	Parent *CommentNode  `json:"parent,omitempty"`
	Items  []CommentNode `json:"items"`
	Meta   PageMeta      `json:"meta"`
	Sort   string        `json:"sort"`
}

type SearchResponse struct {
	Query string    `json:"query"`
	Items []Comment `json:"items"`
	Meta  PageMeta  `json:"meta"`
	Sort  string    `json:"sort"`
}

type CreateCommentRequest struct {
	ParentID int64  `json:"parent_id"`
	Author   string `json:"author"`
	Text     string `json:"text"`
}
