package models

import (
	"time"
)

type User struct {
	UserId    int64     `xorm:"not null pk autoincr BIGINT(20)"`
	Password  string    `xorm:"not null VARCHAR(40)"`
	Name      string    `xorm:"not null VARCHAR(30)"`
	CreatedAt time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' comment('记录创建时间') TIMESTAMP"`
	UpdatedAt time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' comment('记录更新时间') TIMESTAMP"`
}
