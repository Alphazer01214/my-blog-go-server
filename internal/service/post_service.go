package service

import (
	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"

	"gorm.io/gorm"
)

type PostService struct{}

func (ps *PostService) Create(post *entity.Post) (*response.PostDetail, error) {
	if err := ps.create(post); err != nil {
		return nil, err
	}
	author, err := Service.UserService.GetUserInfoById(post.UserId)
	if err != nil {
		return nil, err
	}
	return ps.toPostDetail(post, &author, 0), nil
}

func (ps *PostService) QueryOneById(id uint, viewerId uint) (*response.PostDetail, error) {
	post, err := ps.queryOneById(id)
	if err != nil {
		return nil, err
	}
	author, err := Service.UserService.GetUserInfoById(post.UserId)
	if err != nil {
		return nil, err
	}
	return ps.toPostDetail(post, &author, viewerId), nil
}

func (ps *PostService) QueryAll(page, pageSize int, viewerId uint) (response.PostList, error) {
	var posts []entity.Post
	var total int64
	db := global.GetDB().Model(&entity.Post{})
	if err := db.Count(&total).Error; err != nil {
		return response.PostList{}, err
	}
	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&posts).Error; err != nil {
		return response.PostList{}, err
	}

	items := make([]response.PostDetail, len(posts))
	for i, post := range posts {
		author, err := Service.UserService.GetUserInfoById(post.UserId)
		if err != nil {
			return response.PostList{}, err
		}
		items[i] = *ps.toPostDetail(&post, &author, viewerId)
	}

	return response.PostList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (ps *PostService) DeleteById(id uint) error {
	return global.GetDB().Delete(&entity.Post{}, id).Error
}

func (ps *PostService) Update(id uint, req request.PostUpdateRequest) error {
	dbPost, err := ps.queryOneById(id)
	if err != nil {
		return err
	}

	dbPost.Title = req.Title
	dbPost.Cover = req.Cover
	dbPost.Category = req.Category
	dbPost.Tags = req.Tags
	dbPost.Keywords = req.Keywords
	dbPost.Content = req.Content
	dbPost.Public = req.Public
	return global.GetDB().Save(dbPost).Error
}

func (ps *PostService) Like(postId, userId uint) error {
	return global.GetDB().Transaction(func(tx *gorm.DB) error {
		var existingLike entity.PostLike
		result := tx.Where("post_id = ? AND user_id = ?", postId, userId).First(&existingLike)
		if result.RowsAffected > 0 {
			if err := tx.Delete(&existingLike).Error; err != nil {
				return err
			}
			return tx.Model(&entity.Post{}).Where("id = ?", postId).
				Update("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error
		}

		r := tx.Where("post_id = ? AND user_id = ?", postId, userId).Delete(&entity.PostDislike{})
		if r.RowsAffected > 0 {
			if err := tx.Model(&entity.Post{}).Where("id = ?", postId).
				Update("dislike_count", gorm.Expr("GREATEST(dislike_count - 1, 0)")).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(&entity.PostLike{PostId: postId, Like: entity.Like{UserId: userId}}).Error; err != nil {
			return err
		}
		return tx.Model(&entity.Post{}).Where("id = ?", postId).
			Update("like_count", gorm.Expr("like_count + 1")).Error
	})
}

func (ps *PostService) Dislike(postId, userId uint) error {
	return global.GetDB().Transaction(func(tx *gorm.DB) error {
		var existingDislike entity.PostDislike
		result := tx.Where("post_id = ? AND user_id = ?", postId, userId).First(&existingDislike)
		if result.RowsAffected > 0 {
			if err := tx.Delete(&existingDislike).Error; err != nil {
				return err
			}
			return tx.Model(&entity.Post{}).Where("id = ?", postId).
				Update("dislike_count", gorm.Expr("GREATEST(dislike_count - 1, 0)")).Error
		}

		r := tx.Where("post_id = ? AND user_id = ?", postId, userId).Delete(&entity.PostLike{})
		if r.RowsAffected > 0 {
			if err := tx.Model(&entity.Post{}).Where("id = ?", postId).
				Update("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(&entity.PostDislike{PostId: postId, Dislike: entity.Dislike{UserId: userId}}).Error; err != nil {
			return err
		}
		return tx.Model(&entity.Post{}).Where("id = ?", postId).
			Update("dislike_count", gorm.Expr("dislike_count + 1")).Error
	})
}

func (ps *PostService) GetPostsByUser(userId uint, page, pageSize int) (response.PostList, error) {
	var posts []entity.Post
	var total int64
	db := global.GetDB().Model(&entity.Post{}).Where("user_id = ?", userId)
	if err := db.Count(&total).Error; err != nil {
		return response.PostList{}, err
	}
	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&posts).Error; err != nil {
		return response.PostList{}, err
	}

	author, err := Service.UserService.GetUserInfoById(userId)
	if err != nil {
		return response.PostList{}, err
	}
	items := make([]response.PostDetail, len(posts))
	for i, post := range posts {
		items[i] = *ps.toPostDetail(&post, &author, 0)
	}

	return response.PostList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (ps *PostService) create(post *entity.Post) error {
	return global.GetDB().Create(post).Error
}

func (ps *PostService) queryOneById(id uint) (*entity.Post, error) {
	var post entity.Post
	err := global.GetDB().Where("id = ?", id).First(&post).Error
	return &post, err
}

func (ps *PostService) toPostDetail(post *entity.Post, author *response.UserInfo, viewerId uint) *response.PostDetail {
	pd := &response.PostDetail{
		ID:           post.ID,
		CreatedAt:    post.CreatedAt,
		UpdatedAt:    post.UpdatedAt,
		Title:        post.Title,
		Cover:        post.Cover,
		UserId:       post.UserId,
		Tags:         post.Tags,
		Category:     post.Category,
		Keywords:     post.Keywords,
		Content:      post.Content,
		ViewCount:    post.ViewCount,
		CommentCount: post.CommentCount,
		LikeCount:    post.LikeCount,
		DislikeCount: post.DislikeCount,
		Public:       post.Public,
	}
	if author != nil {
		pd.Author = *author
	}

	if viewerId > 0 {
		var like entity.PostLike
		if global.GetDB().Where("post_id = ? AND user_id = ?", post.ID, viewerId).First(&like).Error == nil {
			pd.IsLiked = true
		}
		var dislike entity.PostDislike
		if global.GetDB().Where("post_id = ? AND user_id = ?", post.ID, viewerId).First(&dislike).Error == nil {
			pd.IsDisliked = true
		}
	}

	return pd
}
