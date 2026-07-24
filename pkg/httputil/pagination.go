package httputil

import (
	"net/http"
	"strconv"
)

type PaginatedResponse struct {
	Items any `json:"items"`
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

func ParsePagination(r *http.Request) (int, int) {
	page := parseIntDefault(r, "page", 1)
	limit := parseIntDefault(r, "limit", 20)

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return page, limit
}

func Offset(page, limit int) int {
	return (page - 1) * limit
}

func parseIntDefault(r *http.Request, key string, fallback int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
