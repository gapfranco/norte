package models

type User struct {
	ID      int
	Usuario string
	Nome    string
}

type PaginationPageItem struct {
	Page     int
	Current  bool
	Ellipsis bool
}

type PaginationMetadata struct {
	CurrentPage int
	PageSize    int
	TotalItems  int
	TotalPages  int
	HasPrev     bool
	HasNext     bool
	PrevPage    int
	NextPage    int
	Pages       []PaginationPageItem
}
