package service

import (
	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
)

type PostService struct {
}

func (ps *PostService) Create(post *entity.Post) error {
	return ps.create(post)
}

func (ps *PostService) QueryOneById(id uint) (*entity.Post, error) {
	return ps.queryOneById(id)
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
