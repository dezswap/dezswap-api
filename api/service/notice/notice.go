package notice

import (
	models "github.com/dezswap/dezswap-api/pkg/db/api"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Notice interface {
	Notices(chain string, cond PaginationCond) ([]NoticeItem, error)
}

type notice struct {
	*gorm.DB
}

var _ Notice = &notice{}

func NewService(db *gorm.DB) Notice {
	return &notice{db}
}

// Notices implements Notice.
func (n *notice) Notices(chain string, cond PaginationCond) ([]NoticeItem, error) {
	cond.Trim()

	query := n.DB.Model(&models.Notice{}).Select("id, title, description, date AT TIME ZONE 'UTC' as date, chain")
	if chain != "" {
		query = query.Where("chain = ?", chain)
	}

	if cond.Limit > 0 {
		query = query.Limit(int(cond.Limit))
	}

	if cond.Asc {
		query = query.Where("id > ?", cond.After).Order("id asc")
	} else {
		query = query.Where("id < ?", cond.After).Order("id desc")
	}

	items := []NoticeItem{}
	if err := query.Find(&items).Error; err != nil {
		return nil, errors.Wrap(err, "notice.Notices")
	}
	return items, nil
}
