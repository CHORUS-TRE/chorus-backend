package model

type Pagination struct {
	Offset uint64
	Limit  uint64
	Sort   Sort
}

type Sort struct {
	SortOrder string
	SortType  string
}
