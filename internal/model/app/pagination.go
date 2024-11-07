package app

import (
	"fmt"
)

type Pagination struct {
	CurrentPage  int
	LenItems     int
	ItemsPerPage int
}

func (p Pagination) MidTitle() string {
	representPage := p.CurrentPage + 1
	return fmt.Sprintf("%d/%d", representPage, p.Pages())
}

func (p Pagination) Pages() int {
	pages := p.LenItems / p.ItemsPerPage
	if p.LenItems%p.ItemsPerPage > 0 {
		return pages + 1
	}
	return pages
}

func (p Pagination) NextTitle() string {
	return "▶"
}

func (p Pagination) PreviousTitle() string {
	return "◀"
}

func (p Pagination) NextPage() int {
	nextPage := p.CurrentPage + 1
	if nextPage >= p.Pages() {
		return 0
	}
	return nextPage
}

func (p Pagination) PrevPage() int {
	prevPage := p.CurrentPage - 1
	if prevPage < 0 {
		prevPage = p.Pages() - 1
	}
	if prevPage < 0 {
		prevPage = 0
	}
	return prevPage
}
