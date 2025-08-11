package store

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// PaginatedFeedQuery represents the query parameters for paginated feed requests.
type PaginatedFeedQuery struct {
	Limit  int      `json:"limit" validate:"gte=1,lte=20"`  // Maximum number of posts to return, between 1 and 20
	Offset int      `json:"offset" validate:"gte=0"`        // Offset for pagination, starting from 0
	Sort   string   `json:"sort" validate:"oneof=asc desc"` // Sort by ascending or descending order
	Tags   []string `json:"tags" validate:"max=5"`          // Optional tags to filter posts
	Search string   `json:"search" validate:"max=100"`      // Optional search term to filter posts
	Since  string   `json:"since"`                          // Optional timestamp to filter posts since a specific date
	Until  string   `json:"until"`                          // Optional timestamp to filter posts until a specific date
}

// Parse extracts the pagination parameters from the HTTP request and returns a PaginatedFeedQuery.
func (fq PaginatedFeedQuery) Parse(r *http.Request) (PaginatedFeedQuery, error) {
	// get the query parameters from the request URL
	qs := r.URL.Query()

	// Parse limit
	limit := qs.Get("limit")
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return fq, err
		}

		fq.Limit = l
	}

	// Parse offset
	offset := qs.Get("offset")
	if offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil {
			return fq, err
		}

		fq.Offset = o
	}

	// Parse sorting order
	sort := qs.Get("sort")
	if sort != "" {
		fq.Sort = sort
	}

	// Parse tags
	tags := qs.Get("tags")
	if tags != "" {
		fq.Tags = strings.Split(tags, ",")
	}

	// Parse search term
	search := qs.Get("search")
	if search != "" {
		fq.Search = search
	}

	// Parse since timestamp
	since := qs.Get("since")
	if since != "" {
		fq.Since = parseTime(since)
	}

	// Parse until timestamp
	until := qs.Get("until")
	if until != "" {
		fq.Until = parseTime(until)
	}

	return fq, nil
}

func parseTime(s string) string {
	t, err := time.Parse(time.DateTime, s)
	if err != nil {
		return ""
	}

	return t.Format(time.DateTime)
}
