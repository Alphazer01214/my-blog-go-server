package service

import (
	"errors"
	"fmt"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/response"
	"gorm.io/gorm"
)

type CommentService struct{}

func (cs *CommentService) Create(comment *entity.Comment) (*response.Comment, error) {
	err := global.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}

		if comment.ParentCommentId != 0 {
			if err := tx.Model(&entity.Comment{}).Where("id = ?", comment.ParentCommentId).
				Update("replies", gorm.Expr("replies + 1")).Error; err != nil {
				return err
			}
		}

		return tx.Model(&entity.Post{}).Where("id = ?", comment.PostId).
			Update("comment_count", gorm.Expr("comment_count + 1")).Error
	})
	if err != nil {
		return nil, err
	}

	r := toResponseComment(comment, nil, nil, nil)
	if author, err := Service.UserService.GetUserInfoById(comment.UserId); err == nil {
		r.Author = author
	}
	return &r, nil
}

func (cs *CommentService) GetById(id uint) (*entity.Comment, error) {
	var comment entity.Comment
	err := global.GetDB().Where("id = ?", id).First(&comment).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (cs *CommentService) QueryByPost(postId uint, page, pageSize int, viewerId uint) ([]response.Comment, int64, error) {
	var roots []entity.Comment
	var total int64
	db := global.GetDB().Model(&entity.Comment{}).Where("post_id = ? AND root_comment_id = 0", postId)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&roots).Error; err != nil {
		return nil, 0, err
	}

	if len(roots) == 0 {
		return []response.Comment{}, total, nil
	}

	rootIds := make([]uint, len(roots))
	for i, r := range roots {
		rootIds[i] = r.ID
	}

	var allReplies []entity.Comment
	if err := global.GetDB().Model(&entity.Comment{}).
		Where("root_comment_id IN ?", rootIds).
		Order("created_at asc").
		Find(&allReplies).Error; err != nil {
		return nil, 0, err
	}

	allComments := append([]entity.Comment{}, roots...)
	allComments = append(allComments, allReplies...)
	authorMap := make(map[uint]response.UserInfo)
	for _, c := range allComments {
		if _, exists := authorMap[c.UserId]; !exists {
			if info, err := Service.UserService.GetUserInfoById(c.UserId); err == nil {
				authorMap[c.UserId] = info
			}
		}
	}

	likedIds := make(map[uint]bool)
	dislikedIds := make(map[uint]bool)
	if viewerId > 0 {
		allIds := make([]uint, len(allComments))
		for i, c := range allComments {
			allIds[i] = c.ID
		}
		var likes []entity.CommentLike
		global.GetDB().Where("comment_id IN ? AND user_id = ?", allIds, viewerId).Find(&likes)
		for _, l := range likes {
			likedIds[l.CommentId] = true
		}
		var dislikes []entity.CommentDislike
		global.GetDB().Where("comment_id IN ? AND user_id = ?", allIds, viewerId).Find(&dislikes)
		for _, d := range dislikes {
			dislikedIds[d.CommentId] = true
		}
	}

	result := make([]response.Comment, len(roots))
	for i, root := range roots {
		rc := toResponseComment(&root, authorMap, likedIds, dislikedIds)
		rc.ReplyComments = buildNestedReplies(root.ID, allReplies, authorMap, likedIds, dislikedIds)
		result[i] = rc
	}
	return result, total, nil
}

func (cs *CommentService) QueryReplies(rootCommentId uint, page, pageSize int, viewerId uint) ([]response.Comment, int64, error) {
	var replies []entity.Comment
	var total int64
	db := global.GetDB().Model(&entity.Comment{}).Where("root_comment_id = ? AND parent_comment_id = ?", rootCommentId, rootCommentId)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	allRepliesDb := global.GetDB().Model(&entity.Comment{}).Where("root_comment_id = ?", rootCommentId)
	if err := allRepliesDb.Order("created_at asc").Find(&replies).Error; err != nil {
		return nil, 0, err
	}

	if len(replies) == 0 {
		return []response.Comment{}, total, nil
	}

	authorMap := make(map[uint]response.UserInfo)
	for _, c := range replies {
		if _, exists := authorMap[c.UserId]; !exists {
			if info, err := Service.UserService.GetUserInfoById(c.UserId); err == nil {
				authorMap[c.UserId] = info
			}
		}
	}

	likedIds := make(map[uint]bool)
	dislikedIds := make(map[uint]bool)
	if viewerId > 0 {
		allIds := make([]uint, len(replies))
		for i, c := range replies {
			allIds[i] = c.ID
		}
		var likes []entity.CommentLike
		global.GetDB().Where("comment_id IN ? AND user_id = ?", allIds, viewerId).Find(&likes)
		for _, l := range likes {
			likedIds[l.CommentId] = true
		}
		var dislikes []entity.CommentDislike
		global.GetDB().Where("comment_id IN ? AND user_id = ?", allIds, viewerId).Find(&dislikes)
		for _, d := range dislikes {
			dislikedIds[d.CommentId] = true
		}
	}

	directReplies := buildNestedReplies(rootCommentId, replies, authorMap, likedIds, dislikedIds)

	offset := (page - 1) * pageSize
	if offset >= len(directReplies) {
		return []response.Comment{}, total, nil
	}
	end := offset + pageSize
	if end > len(directReplies) {
		end = len(directReplies)
	}

	return directReplies[offset:end], total, nil
}

func (cs *CommentService) Delete(commentId, userId uint) error {
	comment, err := cs.GetById(commentId)
	if err != nil {
		return err
	}
	if comment.UserId != userId && comment.RootCommentId == 0 {
		return errors.New("you can't delete others' comment")
	}

	return global.GetDB().Transaction(func(tx *gorm.DB) error {

		if comment.RootCommentId == 0 {
			// 这个用来处理 root comment 的情形
			// 删除所有 root_comment_id 为 commentId 的评论，影响 post 的评论计数
			res := tx.Where("root_comment_id = ?", commentId).Delete(&entity.Comment{})
			delCnt := res.RowsAffected
			if err := tx.Model(&entity.Post{}).Where("id = ?", comment.PostId).Update("comment_count", gorm.Expr(fmt.Sprintf("GREATEST(comment_count - %v)", delCnt))).Error; err != nil {
				return err
			}

			if err := tx.Delete(&entity.Comment{}, commentId).Error; err != nil {
				return err
			}
			return nil
		}

		// 这个用来处理不是 root 的情形
		// 不是 root， 影响 parent， root， post 的计数
		res := tx.Where("parent_comment_id = ?", commentId).Delete(&entity.Comment{})
		delCnt := res.RowsAffected
		// post
		if err := tx.Model(&entity.Post{}).Where("id = ?", comment.PostId).Update("comment_count", gorm.Expr(fmt.Sprintf("GREATEST(comment_count - %v)", delCnt))).Error; err != nil {
			return err
		}
		// root
		if err := tx.Model(&entity.Comment{}).Where("id = ?", comment.RootCommentId).Update("replies", gorm.Expr(fmt.Sprintf("GREATEST(replies - %v)", delCnt))).Error; err != nil {
			return err
		}
		// parent
		if comment.ParentCommentId != 0 {
			if err := tx.Model(&entity.Comment{}).Where("id = ?", comment.ParentCommentId).Update("replies", gorm.Expr(fmt.Sprintf("GREATEST(replies - %v)", delCnt))).Error; err != nil {
				return err
			}
		}

		if err := tx.Delete(&entity.Comment{}, commentId).Error; err != nil {
			return err
		}

		return nil
	})
}

func (cs *CommentService) Like(commentId, userId uint) error {
	return global.GetDB().Transaction(func(tx *gorm.DB) error {
		var existingLike entity.CommentLike

		result := tx.Where("comment_id = ? AND user_id = ?", commentId, userId).First(&existingLike)
		if result.RowsAffected > 0 {
			// 这位用户赞了这个评论
			if err := tx.Delete(&existingLike).Error; err != nil {
				return err
			}
			return tx.Model(&entity.Comment{}).Where("id = ?", commentId).
				Update("likes", gorm.Expr("GREATEST(likes - 1, 0)")).Error
		}

		r := tx.Where("comment_id = ? AND user_id = ?", commentId, userId).Delete(&entity.CommentDislike{})
		if r.RowsAffected > 0 {
			if err := tx.Model(&entity.Comment{}).Where("id = ?", commentId).
				Update("dislikes", gorm.Expr("GREATEST(dislikes - 1, 0)")).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(&entity.CommentLike{CommentId: commentId, Like: entity.Like{UserId: userId}}).Error; err != nil {
			return err
		}
		return tx.Model(&entity.Comment{}).Where("id = ?", commentId).
			Update("likes", gorm.Expr("likes + 1")).Error
	})
}

func (cs *CommentService) Dislike(commentId, userId uint) error {
	return global.GetDB().Transaction(func(tx *gorm.DB) error {
		var existingDislike entity.CommentDislike
		result := tx.Where("comment_id = ? AND user_id = ?", commentId, userId).First(&existingDislike)
		if result.RowsAffected > 0 {
			if err := tx.Delete(&existingDislike).Error; err != nil {
				return err
			}
			return tx.Model(&entity.Comment{}).Where("id = ?", commentId).
				Update("dislikes", gorm.Expr("GREATEST(dislikes - 1, 0)")).Error
		}

		r := tx.Where("comment_id = ? AND user_id = ?", commentId, userId).Delete(&entity.CommentLike{})
		if r.RowsAffected > 0 {
			if err := tx.Model(&entity.Comment{}).Where("id = ?", commentId).
				Update("likes", gorm.Expr("GREATEST(likes - 1, 0)")).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(&entity.CommentDislike{CommentId: commentId, Dislike: entity.Dislike{UserId: userId}}).Error; err != nil {
			return err
		}
		return tx.Model(&entity.Comment{}).Where("id = ?", commentId).
			Update("dislikes", gorm.Expr("dislikes + 1")).Error
	})
}

func toResponseComment(c *entity.Comment, authorMap map[uint]response.UserInfo, likedIds map[uint]bool, dislikedIds map[uint]bool) response.Comment {
	rc := response.Comment{
		CommentId:       c.ID,
		UserId:          c.UserId,
		PostId:          c.PostId,
		RootCommentId:   c.RootCommentId,
		ParentCommentId: c.ParentCommentId,
		Content:         c.Content,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
		Likes:           c.Likes,
		Dislikes:        c.Dislikes,
		Replies:         c.Replies,
		IsLiked:         likedIds[c.ID],
		IsDisliked:      dislikedIds[c.ID],
	}
	if authorMap != nil {
		if author, ok := authorMap[c.UserId]; ok {
			rc.Author = author
		}
	}
	return rc
}

func buildNestedReplies(parentId uint, flatReplies []entity.Comment, authorMap map[uint]response.UserInfo, likedIds map[uint]bool, dislikedIds map[uint]bool) []response.Comment {
	result := make([]response.Comment, 0)
	for i := range flatReplies {
		if flatReplies[i].ParentCommentId == parentId {
			rc := toResponseComment(&flatReplies[i], authorMap, likedIds, dislikedIds)
			rc.ReplyComments = buildNestedReplies(flatReplies[i].ID, flatReplies, authorMap, likedIds, dislikedIds)
			result = append(result, rc)
		}
	}
	return result
}
