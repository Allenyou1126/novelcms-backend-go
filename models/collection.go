package models

import (
	"backend-go/utils"
	"time"
)

type Collection struct {
	ID          int64     `gorm:"column:cid;primaryKey;" json:"cid"`
	AuthorID    int64     `gorm:"column:author" json:"-"`
	Author      User      `gorm:"foreignKey:AuthorID" json:"author"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Articles    []Article `gorm:"foreignKey:CollectionID" json:"-"`
}

var cidGen *utils.IdGenerator = nil

func CreateCollection(author int64, title string) Collection {
	if cidGen == nil {
		g := utils.GetIdGenerator(0, 4, 0)
		cidGen = &g
	}
	return Collection{
		ID:       cidGen.Next(),
		AuthorID: author,
		Title:    title,
	}
}
