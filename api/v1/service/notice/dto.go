package notice

import (
	"time"
)

const DEFAULT_CHAIN = "dimension"
const MAX_LIMIT = 30

var DefaultPaginationCond = PaginationCond{
	Limit: 0,
	Asc:   false,
	After: ^uint(0) >> 1,
}

type NoticeItem struct {
	Id          uint      `json:"id"`
	Chain       string    `json:"chain"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Date        time.Time `json:"date" gorm:"index:notice_date_idx type:timestamp without time zone"`
}

// PaginationCond contains pagination conditions
// Limit is the maximum number of items to return
// Asc is the order of items to return (default: descending order)
type PaginationCond struct {
	After uint `json:"after,omitempty"`
	Limit uint `json:"limit,omitempty"`
	Asc   bool `json:"asc,omitempty"`
}

func (p *PaginationCond) Trim() {
	if p.Limit > MAX_LIMIT {
		p.Limit = MAX_LIMIT
	}

	if p.Asc && p.After == DefaultPaginationCond.After {
		p.After = 0
	}

	if !p.Asc && p.After == 0 {
		p.After = DefaultPaginationCond.After
	}
}
