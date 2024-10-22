package app

import (
	"fmt"
)

type Pagination[T any] struct {
	CurrentPage  int
	ItemsPerPage int
	DataSource   []T
}

func (p *Pagination[T]) MidTitle() string {
	return fmt.Sprintf("%d/%d", p.CurrentPage+1, p.Pages())
}

func (p *Pagination[T]) Len() int {
	return len(p.DataSource)
}

func (p *Pagination[T]) Pages() int {
	pages := p.Len() / p.ItemsPerPage
	if p.Len()%p.ItemsPerPage > 0 {
		pages += 1
	}
	return pages
}

func (p *Pagination[T]) NextTitle() string {
	return "▶"
}

func (p *Pagination[T]) Previous() string {
	return "◀"
}
