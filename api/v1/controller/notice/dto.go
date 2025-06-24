package notice

import (
	"fmt"
	"github.com/dezswap/dezswap-api/api/v1/service/notice"
	"time"
)

type NoticesRes []noticeItem
type noticeItem struct {
	Id          string    `json:"id"`
	Chain       string    `json:"chain"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

type PaginationReq struct {
	Chain string `form:"chain"`
	After uint   `form:"after"`
	Limit uint   `form:"limit"`
	Asc   bool   `form:"asc"`
}

func (p PaginationReq) Default() PaginationReq {
	d := notice.DefaultPaginationCond
	if !p.Asc && p.After == 0 {
		p.After = d.After
	}
	if p.Limit == 0 {
		p.Limit = d.Limit
	}
	return p
}

func (p PaginationReq) ToCondition() notice.PaginationCond {
	return notice.PaginationCond{
		After: p.After,
		Limit: p.Limit,
		Asc:   p.Asc,
	}
}

type mapper struct{}

func (m *mapper) noticesToRes(notices []notice.NoticeItem) NoticesRes {
	res := make(NoticesRes, len(notices))
	for i, n := range notices {
		res[i] = noticeItem{
			Id:          fmt.Sprint(n.Id),
			Chain:       n.Chain,
			Title:       n.Title,
			Description: n.Description,
			Timestamp:   n.Date,
		}
	}
	return res
}
