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

	//	updates := map[string]interface{}{
	//		"title":    post.Title,
	//		"category": post.Category,
	//		"content":  post.Content,
	//		"cover":    post.Cover,
	//		"Keywords": post.Keywords,
	//	}
	dbPost, err := ps.queryOneById(id)
	if err != nil {
		return err
	}

	dbPost.Content = post.Content
	dbPost.Category = post.Category
	dbPost.Title = post.Title
	dbPost.Cover = post.Cover
	dbPost.Keywords = post.Keywords
	return global.GetDB().Save(dbPost).Error
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
