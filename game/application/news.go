package application

import (
	"fmt"
	"realm/models"
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
