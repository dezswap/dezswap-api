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

	query := n.DB.Model(&models.Notice{}).Where("chain = ?", chain).Select("id, title, description, date")
	if cond.Asc {
		query = query.Where("id > ?", cond.After).Limit(int(cond.Limit)).Order("id asc")
	} else {
		query = query.Where("id < ?", cond.After).Limit(int(cond.Limit)).Order("id desc")
	}

	items := []NoticeItem{}
	if err := query.Find(&items).Error; err != nil {
		return nil, errors.Wrap(err, "notice.Notices")
	}
	return items, nil
}
