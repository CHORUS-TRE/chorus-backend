package model

// Constants for pagination and sorting
const DEFAULT_LIMIT = 20 // Default number of items to return if not specified
const MAX_LIMIT = 100    // Maximum number of items to return in a single query

type Sortable interface {
	IsValidSortType(sortType string) bool
}

type Pagination struct {
	Offset uint64
	Limit  uint64
	Sort   Sort
}

type Sort struct {
	SortOrder string
	SortType  string
}

type PaginationResult struct {
	Total  uint64
	Offset uint64
	Limit  uint64
	Sort   Sort
}
