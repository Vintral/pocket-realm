package models

type UserLog struct {
	BaseModel

	UserID  uint   `json:"-"`
	RoundID uint   `json:"round"`
	Message string `json:"message"`
}
