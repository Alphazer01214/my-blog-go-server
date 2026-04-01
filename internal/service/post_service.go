package service

import (
	"errors"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
)

type PostService struct{}

func (ps *PostService) Create(post *entity.Post) error {
	return ps.create(post)
}

func (ps *PostService) QueryOneById(id uint) (*entity.Post, error) {
	return ps.queryOneById(id)
}

func (ps *PostService) Update(id uint, post *entity.Post) error {
	oldPost, err := ps.queryOneById(id)
	if err != nil {
		return err
	}
	if post.AuthorId != oldPost.AuthorId {
		return errors.New("permisstion denied")
	}

	updates := map[string]interface{}{
		"title":    post.Title,
		"category": post.Category,
		"content":  post.Content,
		"cover":    post.Cover,
		"Keywords": post.Keywords,
	}
}

func (ps *PostService) create(post *entity.Post) error {
	return global.GetDB().Create(post).Error
}

func (ps *PostService) queryOneById(id uint) (*entity.Post, error) {
	var post entity.Post
	err := global.GetDB().First(&post, id).Error
	return &post, err
}

//func (ps *PostService) syncToDB(post *entity.Post) error {
//	return global.GetDB().Save(post).Error
//}
