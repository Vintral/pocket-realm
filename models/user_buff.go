package models

import (
	"time"
)

type UserBuff struct {
	BaseModel

	UserID  uint      `json:"-"`
	RoundID uint      `json:"-"`
	BuffID  uint      `json:"-"`
	Stacks  uint      `json:"-"`
	Expires time.Time `json:"-"`
}
