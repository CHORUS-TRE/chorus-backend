package model

type Pagination struct {
	Offset uint64
	Limit  uint64
	Sort   Sort
	Query  map[string][]string
}

type Sort struct {
	SortOrder string
	SortType  string
}
