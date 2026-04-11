package service

import (
	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/request"
)

type PostService struct{}

func (ps *PostService) Create(post *entity.Post) error {
	return ps.create(post)
}

func (ps *PostService) QueryOneById(id uint) (*entity.Post, error) {
	return ps.queryOneById(id)
}

func (ps *PostService) QueryAll() ([]entity.Post, error) {
	return ps.queryAllPostInstance()
}

func (ps *PostService) Update(id uint, req request.PostUpdateRequest) error {

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

	dbPost.Content = req.Content
	dbPost.Category = req.Category
	dbPost.Title = req.Title
	dbPost.Cover = req.Cover
	dbPost.Keywords = req.Keywords
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

func (ps *PostService) queryAllPostInstance() ([]entity.Post, error) {
	var posts []entity.Post
	if err := global.GetDB().Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

//func (ps *PostService) syncToDB(post *entity.Post) error {
//	return global.GetDB().Save(post).Error
//}
