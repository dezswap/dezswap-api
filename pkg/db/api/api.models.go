package api

import (
	"time"

	"gorm.io/gorm"
)

type Notice struct {
	*gorm.Model
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Date        time.Time `json:"date" gorm:"index:notice_date_idx type:timestamp without time zone"`
	Chain       string    `json:"chain" gorm:"not null;index:;comment:chain name of network"`
}

func (n *Notice) TableName() string {
	return "notices"
}
