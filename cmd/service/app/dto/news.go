package dto

import "ago_goredis/pkg/news"

type NewsDTO struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Text    string `json:"text"`
	Created int64  `json:"created"`
}

func FromModel(news *news.News) *NewsDTO {
	return &NewsDTO{
		Id:      news.Id,
		Title:   news.Title,
		Text:    news.Text,
		Created: news.Created,
	}
}