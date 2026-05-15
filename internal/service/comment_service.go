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
				Update("reply_count", gorm.Expr("reply_count + 1")).Error; err != nil {
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
	if author, err := Service.UserService.GetUserInfoById(comment.UserId, 0); err == nil {
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
			if info, err := Service.UserService.GetUserInfoById(c.UserId, 0); err == nil {
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
		var likes []entity.Action
		global.GetDB().Where("target_id IN ? AND user_id = ? AND target_type = ? AND action_type = ?",
			allIds, viewerId, entity.TargetComment, entity.ActionLike).Find(&likes)
		for _, l := range likes {
			likedIds[l.TargetId] = true
		}
		var dislikes []entity.Action
		global.GetDB().Where("target_id IN ? AND user_id = ? AND target_type = ? AND action_type = ?",
			allIds, viewerId, entity.TargetComment, entity.ActionDislike).Find(&dislikes)
		for _, d := range dislikes {
			dislikedIds[d.TargetId] = true
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
			if info, err := Service.UserService.GetUserInfoById(c.UserId, 0); err == nil {
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
		var likes []entity.Action
		global.GetDB().Where("target_id IN ? AND user_id = ? AND target_type = ? AND action_type = ?",
			allIds, viewerId, entity.TargetComment, entity.ActionLike).Find(&likes)
		for _, l := range likes {
			likedIds[l.TargetId] = true
		}
		var dislikes []entity.Action
		global.GetDB().Where("target_id IN ? AND user_id = ? AND target_type = ? AND action_type = ?",
			allIds, viewerId, entity.TargetComment, entity.ActionDislike).Find(&dislikes)
		for _, d := range dislikes {
			dislikedIds[d.TargetId] = true
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
		if err := tx.Model(&entity.Comment{}).Where("id = ?", comment.RootCommentId).Update("reply_count", gorm.Expr(fmt.Sprintf("GREATEST(reply_count - %v)", delCnt))).Error; err != nil {
			return err
		}
		// parent
		if comment.ParentCommentId != 0 {
			if err := tx.Model(&entity.Comment{}).Where("id = ?", comment.ParentCommentId).Update("reply_count", gorm.Expr(fmt.Sprintf("GREATEST(reply_count - %v)", delCnt))).Error; err != nil {
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
		var existing entity.Action

		result := tx.Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			commentId, userId, entity.TargetComment, entity.ActionLike).First(&existing)
		if result.RowsAffected > 0 {
			if err := tx.Delete(&existing).Error; err != nil {
				return err
			}
			return tx.Model(&entity.Comment{}).Where("id = ?", commentId).
				Update("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error
		}

		r := tx.Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			commentId, userId, entity.TargetComment, entity.ActionDislike).Delete(&entity.Action{})
		if r.RowsAffected > 0 {
			if err := tx.Model(&entity.Comment{}).Where("id = ?", commentId).
				Update("dislike_count", gorm.Expr("GREATEST(dislike_count - 1, 0)")).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(&entity.Action{
			UserId:     userId,
			TargetId:   commentId,
			ActType: entity.ActionLike,
			TgtType: entity.TargetComment,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&entity.Comment{}).Where("id = ?", commentId).
			Update("like_count", gorm.Expr("like_count + 1")).Error
	})
}

func (cs *CommentService) Dislike(commentId, userId uint) error {
	return global.GetDB().Transaction(func(tx *gorm.DB) error {
		var existing entity.Action
		result := tx.Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			commentId, userId, entity.TargetComment, entity.ActionDislike).First(&existing)
		if result.RowsAffected > 0 {
			if err := tx.Delete(&existing).Error; err != nil {
				return err
			}
			return tx.Model(&entity.Comment{}).Where("id = ?", commentId).
				Update("dislike_count", gorm.Expr("GREATEST(dislike_count - 1, 0)")).Error
		}

		r := tx.Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			commentId, userId, entity.TargetComment, entity.ActionLike).Delete(&entity.Action{})
		if r.RowsAffected > 0 {
			if err := tx.Model(&entity.Comment{}).Where("id = ?", commentId).
				Update("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(&entity.Action{
			UserId:     userId,
			TargetId:   commentId,
			ActType: entity.ActionDislike,
			TgtType: entity.TargetComment,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&entity.Comment{}).Where("id = ?", commentId).
			Update("dislike_count", gorm.Expr("dislike_count + 1")).Error
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
		Likes:           c.LikeCount,
		Dislikes:        c.DislikeCount,
		Replies:         c.ReplyCount,
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
