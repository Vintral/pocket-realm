package models

import "context"

type Rule struct {
	BaseModel

	Rule   string `json:"rule"`
	Active bool   `json:"-"`
}

func (rule *Rule) Load() []*Rule {
	var rules []*Rule

	ctx, span := Tracer.Start(context.Background(), "rules")
	defer span.End()

	db.WithContext(ctx).Where("active = ?", true).Find(&rules)
	return rules
}
