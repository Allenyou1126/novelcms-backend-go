package models

import (
	"backend-go/utils"
	"time"
)

type Article struct {
	ID           int64     `gorm:"column:aid;primaryKey;" json:"aid"`
	AuthorID     int64     `gorm:"column:author" json:"-"`
	Author       User      `gorm:"foreignKey:AuthorID" json:"author"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	CollectionID int64     `json:"collection_id,omitempty"`
}

var aidGen *utils.IdGenerator = nil

func CreateArticle(author int64, title string, content string) Article {
	if aidGen == nil {
		g := utils.GetIdGenerator(0, 4, 0)
		aidGen = &g
	}
	return Article{
		ID:       aidGen.Next(),
		AuthorID: author,
		Title:    title,
		Content:  content,
	}
}
