package entity

import "gorm.io/gorm"

type Post struct {
	gorm.Model
	EnvInfo  `json:"env"`
	Title    string `json:"title"`
	Cover    string `json:"cover"`
	AuthorId uint   `json:"author_id"`
	Category string `json:"category"`
	Keywords string `json:"keywords"`
	Content  string `json:"content"`

	Views    int `json:"views"`
	Comments int `json:"comments"`
	Likes    int `json:"likes"`
	Dislikes int `json:"dislikes"`

	Public bool `json:"public"`
}
