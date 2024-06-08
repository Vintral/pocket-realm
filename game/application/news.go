package application

import (
	"fmt"

	"github.com/Vintral/pocket-realm-test-access/models"
)

func GetNews(user *models.User) {
	fmt.Println("Get News")

	var newsItems *models.NewsItem
	payload := struct {
		Type      string             `json:"type"`
		NewsItems []*models.NewsItem `json:"news"`
	}{
		Type:      "NEWS",
		NewsItems: newsItems.Load(),
	}
	user.Connection.WriteJSON(payload)
}
