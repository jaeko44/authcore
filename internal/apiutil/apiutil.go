package apiutil

import (
	"authcore.io/authcore/pkg/paging"
)

// ListPagination is a REST API response that represents a paginated listing of a collection.
type ListPagination struct {
	TotalSize     *uint       `json:"total_size"`
	NextPageToken string      `json:"next_page_token,omitempty"`
	PrevPageToken string      `json:"prev_page_token,omitempty"`
	Results       interface{} `json:"results"`
}

// NewListPagination returns a new ListPagination
func NewListPagination(results interface{}, page *paging.Page) *ListPagination {
	lp := &ListPagination{Results: results}
	if page != nil {
		lp.TotalSize = &page.FoundRows
		lp.NextPageToken = page.NextPageToken.String()
		lp.PrevPageToken = page.PreviousPageToken.String()
	}
	return lp
}
