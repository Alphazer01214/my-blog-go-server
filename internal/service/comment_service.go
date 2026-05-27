package service

import (
	"errors"
	"fmt"

	"blog.alphazer01214.top/internal/constant"
	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/response"
	"gorm.io/gorm"
)

type CommentService struct{}

func (cs *CommentService) Create(comment *entity.Comment) (*response.Comment, error) {
	if comment.TargetType == constant.TargetPost && cs.isPostForbidComment(comment.TargetId) {
		return nil, errors.New("this post is forbidden to comment")
	}
	if comment.TargetType == constant.TargetVideo && Service.VideoService.IsVideoForbidComment(comment.TargetId) {
		return nil, errors.New("this video is forbidden to comment")
	}
	err := global.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}

		if err := tx.Model(&entity.UserProfile{}).Where("user_id = ?", comment.UserId).Update("comment_count", gorm.Expr("comment_count + 1")).Error; err != nil {
			return err
		}

		if comment.RootCommentId != 0 {
			if err := tx.Model(&entity.Comment{}).Where("id = ?", comment.RootCommentId).
				Update("reply_count", gorm.Expr("reply_count + 1")).Error; err != nil {
				return err
			}
		}
		if comment.ParentCommentId != 0 && comment.ParentCommentId != comment.RootCommentId {
			if err := tx.Model(&entity.Comment{}).Where("id = ?", comment.ParentCommentId).
				Update("reply_count", gorm.Expr("reply_count + 1")).Error; err != nil {
				return err
			}
		}

		if comment.TargetType == constant.TargetPost {
			return tx.Model(&entity.Post{}).Where("id = ?", comment.TargetId).
				Update("comment_count", gorm.Expr("comment_count + 1")).Error
		}
		if comment.TargetType == constant.TargetVideo {
			return tx.Model(&entity.Video{}).Where("id = ?", comment.TargetId).
				Update("comment_count", gorm.Expr("comment_count + 1")).Error
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	r := cs.toResponseComment(comment, 0)
	if author, err := Service.UserService.GetUserInfoById(comment.UserId, 0); err == nil {
		r.Author = author
	}
	return r, nil
}

func (cs *CommentService) Delete(commentId, userId uint) error {
	comment, err := cs.getEntityByCommentId(commentId)
	if err != nil {
		return err
	}
	if comment.UserId != userId {
		return errors.New("you can't delete others' comment")
	}

	return global.GetDB().Transaction(func(tx *gorm.DB) error {

		if comment.RootCommentId == 0 {
			res := tx.Where("root_comment_id = ?", commentId).Delete(&entity.Comment{})
			delCnt := res.RowsAffected
			if comment.TargetType == constant.TargetPost {
				if err := tx.Model(&entity.Post{}).Where("id = ?", comment.TargetId).Update("comment_count", gorm.Expr("GREATEST(comment_count - ?, 0)", delCnt)).Error; err != nil {
					return err
				}
			}
			if comment.TargetType == constant.TargetVideo {
				if err := tx.Model(&entity.Video{}).Where("id = ?", comment.TargetId).Update("comment_count", gorm.Expr("GREATEST(comment_count - ?, 0)", delCnt)).Error; err != nil {
					return err
				}
			}

			if err := tx.Delete(&entity.Comment{}, commentId).Error; err != nil {
				return err
			}
			return tx.Model(&entity.UserProfile{}).Where("user_id = ?", comment.UserId).Update("comment_count", gorm.Expr("GREATEST(comment_count - 1, 0)")).Error
		}

		res := tx.Where("parent_comment_id = ?", commentId).Delete(&entity.Comment{})
		delCnt := res.RowsAffected
		if comment.TargetType == constant.TargetPost {
			if err := tx.Model(&entity.Post{}).Where("id = ?", comment.TargetId).Update("comment_count", gorm.Expr("GREATEST(comment_count - ?, 0)", delCnt)).Error; err != nil {
				return err
			}
		}
		if comment.TargetType == constant.TargetVideo {
			if err := tx.Model(&entity.Video{}).Where("id = ?", comment.TargetId).Update("comment_count", gorm.Expr("GREATEST(comment_count - ?, 0)", delCnt)).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&entity.Comment{}).Where("id = ?", comment.RootCommentId).Update("reply_count", gorm.Expr("GREATEST(reply_count - ?, 0)", delCnt)).Error; err != nil {
			return err
		}
		if comment.ParentCommentId != 0 && comment.ParentCommentId != comment.RootCommentId {
			if err := tx.Model(&entity.Comment{}).Where("id = ?", comment.ParentCommentId).Update("reply_count", gorm.Expr("GREATEST(reply_count - ?, 0)", delCnt)).Error; err != nil {
				return err
			}
		}

		if err := tx.Delete(&entity.Comment{}, commentId).Error; err != nil {
			return err
		}

		return tx.Model(&entity.UserProfile{}).Where("user_id = ?", comment.UserId).Update("comment_count", gorm.Expr("GREATEST(comment_count - 1, 0)")).Error
	})
}

func (cs *CommentService) Like(commentId, userId uint) error {
	return global.GetDB().Transaction(func(tx *gorm.DB) error {
		var existing entity.Action

		result := tx.Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			commentId, userId, constant.TargetComment, constant.ActionLike).First(&existing)
		if result.RowsAffected > 0 {
			if err := tx.Delete(&existing).Error; err != nil {
				return err
			}
			return tx.Model(&entity.Comment{}).Where("id = ?", commentId).
				Update("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error
		}

		r := tx.Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			commentId, userId, constant.TargetComment, constant.ActionDislike).Delete(&entity.Action{})
		if r.RowsAffected > 0 {
			if err := tx.Model(&entity.Comment{}).Where("id = ?", commentId).
				Update("dislike_count", gorm.Expr("GREATEST(dislike_count - 1, 0)")).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(&entity.Action{
			UserId:   userId,
			TargetId: commentId,
			ActType:  constant.ActionLike,
			TgtType:  constant.TargetComment,
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
			commentId, userId, constant.TargetComment, constant.ActionDislike).First(&existing)
		if result.RowsAffected > 0 {
			if err := tx.Delete(&existing).Error; err != nil {
				return err
			}
			return tx.Model(&entity.Comment{}).Where("id = ?", commentId).
				Update("dislike_count", gorm.Expr("GREATEST(dislike_count - 1, 0)")).Error
		}

		r := tx.Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			commentId, userId, constant.TargetComment, constant.ActionLike).Delete(&entity.Action{})
		if r.RowsAffected > 0 {
			if err := tx.Model(&entity.Comment{}).Where("id = ?", commentId).
				Update("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(&entity.Action{
			UserId:   userId,
			TargetId: commentId,
			ActType:  constant.ActionDislike,
			TgtType:  constant.TargetComment,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&entity.Comment{}).Where("id = ?", commentId).
			Update("dislike_count", gorm.Expr("dislike_count + 1")).Error
	})
}

func (cs *CommentService) GetCommentsByPostId(postId uint, page int, pageSize int, viewerId uint) ([]*response.Comment, int64, error) {
	if cs.isPostForbidComment(postId) {
		return nil, 0, errors.New("this post is forbidden to comment")
	}
	return cs.getCommentsByTarget(constant.TargetPost, postId, page, pageSize, viewerId)
}

func (cs *CommentService) GetCommentsByVideoId(videoId uint, page int, pageSize int, viewerId uint) ([]*response.Comment, int64, error) {
	if Service.VideoService.IsVideoForbidComment(videoId) {
		return nil, 0, errors.New("this video is forbidden to comment")
	}
	return cs.getCommentsByTarget(constant.TargetVideo, videoId, page, pageSize, viewerId)
}

func (cs *CommentService) getCommentsByTarget(targetType constant.TargetType, targetId uint, page int, pageSize int, viewerId uint) ([]*response.Comment, int64, error) {
	var roots []*entity.Comment
	var comments []*response.Comment
	var total int64
	db := global.GetDB().Model(&entity.Comment{}).Where("target_id = ? AND target_type = ? AND root_comment_id = 0", targetId, targetType)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&roots).Error; err != nil {
		return nil, 0, err
	}

	for _, root := range roots {
		r := cs.toResponseComment(root, viewerId)
		r.ReplyComments = cs.getRepliesRecursive(r.CommentId, r.CommentId, viewerId)
		fmt.Printf("[comment service] %v \n", *r)
		comments = append(comments, r)
	}

	return comments, total, nil
}

func (cs *CommentService) GetCommentsByUserId(userId uint, page int, pageSize int) ([]*response.Comment, int64, error) {
	var comments []*entity.Comment
	var res []*response.Comment
	var total int64
	db := global.GetDB().Model(&entity.Comment{}).Where("user_id = ?", userId)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&comments).Error; err != nil {
		return nil, 0, err
	}
	for _, c := range comments {
		res = append(res, cs.toResponseComment(c, userId))
	}

	return res, total, nil
}

func (cs *CommentService) getRepliesRecursive(rootId uint, parentId uint, viewerId uint) []*response.Comment {
	var children []*entity.Comment
	var res []*response.Comment

	if err := global.GetDB().Where("root_comment_id = ? AND parent_comment_id = ?", rootId, parentId).Find(&children).Error; err != nil {
		return nil
	}

	for _, child := range children {
		rep := cs.toResponseComment(child, viewerId)
		rep.ReplyComments = cs.getRepliesRecursive(rootId, rep.CommentId, viewerId)
		res = append(res, rep)
	}
	return res
}

func (cs *CommentService) getEntityByCommentId(id uint) (*entity.Comment, error) {
	var comment *entity.Comment
	if err := global.GetDB().Where("id = ?", id).First(&comment).Error; err != nil {
		return nil, err
	}
	return comment, nil
}

func (cs *CommentService) isLiked(commentId uint, viewerId uint) bool {
	if viewerId == 0 {
		return false
	}
	var action entity.Action
	err := global.GetDB().Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
		commentId, viewerId, constant.TargetComment, constant.ActionLike).First(&action).Error
	return err == nil
}

func (cs *CommentService) isDisliked(commentId uint, viewerId uint) bool {
	if viewerId == 0 {
		return false
	}
	var action entity.Action
	err := global.GetDB().Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
		commentId, viewerId, constant.TargetComment, constant.ActionDislike).First(&action).Error
	return err == nil
}

func (cs *CommentService) isPostForbidComment(postId uint) bool {
	var res bool
	err := global.GetDB().Model(&entity.Post{}).Where("ID = ?", postId).Pluck("forbid_comment", &res).Error
	return err == nil && res
}

//type Comment struct {
//	ID     uint `gorm:"primaryKey" json:"id"`
//	UserId uint `json:"user_id"`
//	PostId uint `json:"post_id"`
//	// RootCommentId 根评论id，如果为0，则表示该评论为根评论
//	RootCommentId   uint           `json:"root_comment_id"`
//	ParentCommentId uint           `json:"parent_comment_id"`
//	CreatedAt       time.Time      `json:"created_at"`
//	UpdatedAt       time.Time      `json:"updated_at"`
//	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
//	EnvInfo         `json:"env"`
//
//	Content string `json:"content"`
//
//	LikeCount    int `json:"like_count"`
//	DislikeCount int `json:"dislike_count"`
//	ReplyCount   int `json:"reply_count"`
//}

//	type Comment struct {
//		CommentId       uint      `json:"comment_id"`
//		UserId          uint      `json:"user_id"`
//		PostId          uint      `json:"post_id"`
//		RootCommentId   uint      `json:"root_comment_id"`
//		ParentCommentId uint      `json:"parent_comment_id"`
//		Content         string    `json:"content"`
//		CreatedAt       time.Time `json:"created_at"`
//		UpdatedAt       time.Time `json:"updated_at"`
//
//		Author UserInfo `json:"author,omitempty"`
//
//		LikeCount    int `json:"likes"`
//		DislikeCount int `json:"dislikes"`
//		ReplyCount  int `json:"replies"`
//
//		IsLiked    bool `json:"is_liked"`
//		IsDisliked bool `json:"is_disliked"`
//
//		ReplyComments []*Comment `json:"reply_comments"`
//	}
//
// toResponseComment 转换为响应 缺少 reply comment
func (cs *CommentService) toResponseComment(comment *entity.Comment, viewerId uint) *response.Comment {
	userInfo, _ := Service.UserService.GetUserInfoById(comment.UserId, viewerId)
	return &response.Comment{
		Author:          userInfo,
		CommentId:       comment.ID,
		Content:         comment.Content,
		CreatedAt:       comment.CreatedAt,
		DislikeCount:    comment.DislikeCount,
		IsDisliked:      cs.isDisliked(comment.ID, viewerId),
		IsLiked:         cs.isLiked(comment.ID, viewerId),
		LikeCount:       comment.LikeCount,
		ReplyCount:      comment.ReplyCount,
		ParentCommentId: comment.ParentCommentId,
		TargetId:        comment.TargetId,
		TargetType:      string(comment.TargetType),
		RootCommentId:   comment.RootCommentId,
		UpdatedAt:       comment.UpdatedAt,
		Env:             comment.EnvInfo,
	}
}
