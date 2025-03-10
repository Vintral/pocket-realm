package models

type UserDevotion struct {
	BaseModel

	UserID   uint `json:"-"`
	RoundID  uint `json:"-"`
	Pantheon uint `json:"-"`
	Level    uint `json:"-"`
}
