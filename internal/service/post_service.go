package service

import (
	"encoding/json"

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
	author, err := Service.UserService.GetUserInfoById(post.UserId, 0)
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
	author, err := Service.UserService.GetUserInfoById(post.UserId, 0)
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
	author, err := Service.UserService.GetUserInfoById(post.UserId, 0)
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
	dbPost.CategoryId = req.CategoryId
	dbPost.Tags = req.Tags
	dbPost.Keywords = req.Keywords
	dbPost.Content = req.Content
	dbPost.Public = req.Public
	return global.GetDB().Save(dbPost).Error
}

func (ps *PostService) Like(postId, userId uint) error {
	return global.GetDB().Transaction(func(tx *gorm.DB) error {
		var existing entity.Action
		result := tx.Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			postId, userId, entity.TargetPost, entity.ActionLike).First(&existing)
		if result.RowsAffected > 0 {
			if err := tx.Delete(&existing).Error; err != nil {
				return err
			}
			return tx.Model(&entity.Post{}).Where("id = ?", postId).
				Update("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error
		}

		r := tx.Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			postId, userId, entity.TargetPost, entity.ActionDislike).Delete(&entity.Action{})
		if r.RowsAffected > 0 {
			if err := tx.Model(&entity.Post{}).Where("id = ?", postId).
				Update("dislike_count", gorm.Expr("GREATEST(dislike_count - 1, 0)")).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(&entity.Action{
			UserId:     userId,
			TargetId:   postId,
			ActType: entity.ActionLike,
			TgtType: entity.TargetPost,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&entity.Post{}).Where("id = ?", postId).
			Update("like_count", gorm.Expr("like_count + 1")).Error
	})
}

func (ps *PostService) Dislike(postId, userId uint) error {
	return global.GetDB().Transaction(func(tx *gorm.DB) error {
		var existing entity.Action
		result := tx.Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			postId, userId, entity.TargetPost, entity.ActionDislike).First(&existing)
		if result.RowsAffected > 0 {
			if err := tx.Delete(&existing).Error; err != nil {
				return err
			}
			return tx.Model(&entity.Post{}).Where("id = ?", postId).
				Update("dislike_count", gorm.Expr("GREATEST(dislike_count - 1, 0)")).Error
		}

		r := tx.Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			postId, userId, entity.TargetPost, entity.ActionLike).Delete(&entity.Action{})
		if r.RowsAffected > 0 {
			if err := tx.Model(&entity.Post{}).Where("id = ?", postId).
				Update("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(&entity.Action{
			UserId:     userId,
			TargetId:   postId,
			ActType: entity.ActionDislike,
			TgtType: entity.TargetPost,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&entity.Post{}).Where("id = ?", postId).
			Update("dislike_count", gorm.Expr("dislike_count + 1")).Error
	})
}

func (ps *PostService) Favorite(postId, userId uint) error {
	return global.GetDB().Transaction(func(tx *gorm.DB) error {
		var existing entity.Action
		result := tx.Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			postId, userId, entity.TargetPost, entity.ActionFavorite).First(&existing)
		if result.RowsAffected > 0 {
			if err := tx.Delete(&existing).Error; err != nil {
				return err
			}
			return tx.Model(&entity.Post{}).Where("id = ?", postId).
				Update("favorite_count", gorm.Expr("GREATEST(favorite_count - 1, 0)")).Error
		}

		if err := tx.Create(&entity.Action{
			UserId:     userId,
			TargetId:   postId,
			ActType: entity.ActionFavorite,
			TgtType: entity.TargetPost,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&entity.Post{}).Where("id = ?", postId).
			Update("favorite_count", gorm.Expr("favorite_count + 1")).Error
	})
}

func (ps *PostService) Share(postId, userId uint, shareInfo entity.ShareInfo) error {
	extraJSON, err := json.Marshal(shareInfo)
	if err != nil {
		return err
	}
	return global.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&entity.Action{
			UserId:     userId,
			TargetId:   postId,
			ActType: entity.ActionShare,
			TgtType: entity.TargetPost,
			ExtraInfo:  extraJSON,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&entity.Post{}).Where("id = ?", postId).
			Update("share_count", gorm.Expr("share_count + 1")).Error
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

	author, err := Service.UserService.GetUserInfoById(userId, 0)
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
	var categoryName string
	if post.CategoryId != 0 {
		var cat entity.Category
		if err := global.GetDB().Where("id = ?", post.CategoryId).First(&cat).Error; err == nil {
			categoryName = cat.Name
		}
	}

	pd := &response.PostDetail{
		ID:            post.ID,
		CreatedAt:     post.CreatedAt,
		UpdatedAt:     post.UpdatedAt,
		Title:         post.Title,
		Cover:         post.Cover,
		UserId:        post.UserId,
		Tags:          post.Tags,
		CategoryId:    post.CategoryId,
		CategoryName:  categoryName,
		Keywords:      post.Keywords,
		Content:       post.Content,
		ViewCount:     post.ViewCount,
		CommentCount:  post.CommentCount,
		LikeCount:     post.LikeCount,
		DislikeCount:  post.DislikeCount,
		FavoriteCount: post.FavoriteCount,
		ShareCount:    post.ShareCount,
		Public:        post.Public,
	}
	if author != nil {
		pd.Author = *author
	}

	if viewerId > 0 {
		var like entity.Action
		if global.GetDB().Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			post.ID, viewerId, entity.TargetPost, entity.ActionLike).First(&like).Error == nil {
			pd.IsLiked = true
		}
		var dislike entity.Action
		if global.GetDB().Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			post.ID, viewerId, entity.TargetPost, entity.ActionDislike).First(&dislike).Error == nil {
			pd.IsDisliked = true
		}
		var fav entity.Action
		if global.GetDB().Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			post.ID, viewerId, entity.TargetPost, entity.ActionFavorite).First(&fav).Error == nil {
			pd.IsFavorited = true
		}
	}

	return pd
}
