package application

import (
	"fmt"

	"github.com/Vintral/pocket-realm/models"
)

func GetRules(user *models.User) {
	fmt.Println("Get Rules")

	var rule *models.Rule
	payload := struct {
		Type  string         `json:"type"`
		Rules []*models.Rule `json:"rules"`
	}{
		Type:  "RULES",
		Rules: rule.Load(),
	}
	user.Connection.WriteJSON(payload)
}
