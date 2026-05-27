package service

import (
	"context"
	"errors"

	"blog.alphazer01214.top/internal/constant"
	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/repository"
	"blog.alphazer01214.top/internal/response"

	"gorm.io/gorm"
)

type CommentService struct {
	commentRepo repository.CommentRepository
	postRepo    repository.PostRepository
	videoRepo   repository.VideoRepository
	userRepo    repository.UserRepository
	actionRepo  repository.ActionRepository
	userService *UserService
}

func NewCommentService(
	commentRepo repository.CommentRepository,
	postRepo repository.PostRepository,
	videoRepo repository.VideoRepository,
	userRepo repository.UserRepository,
	actionRepo repository.ActionRepository,
	userService *UserService,
) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		postRepo:    postRepo,
		videoRepo:   videoRepo,
		userRepo:    userRepo,
		actionRepo:  actionRepo,
		userService: userService,
	}
}

func (cs *CommentService) Create(ctx context.Context, comment *entity.Comment) (*response.Comment, error) {
	if comment.TargetType == constant.TargetPost {
		forbidden, _ := cs.postRepo.GetForbidComment(ctx, comment.TargetId)
		if forbidden {
			return nil, errors.New("this post is forbidden to comment")
		}
	}
	if comment.TargetType == constant.TargetVideo {
		forbidden, _ := cs.videoRepo.GetForbidComment(ctx, comment.TargetId)
		if forbidden {
			return nil, errors.New("this video is forbidden to comment")
		}
	}

	err := cs.commentRepo.WithTransaction(ctx, func(tx *gorm.DB) error {
		txComment := cs.commentRepo.WithTx(tx)
		txUser := cs.userRepo.WithTx(tx)
		txPost := cs.postRepo.WithTx(tx)
		txVideo := cs.videoRepo.WithTx(tx)

		if err := txComment.Create(ctx, comment); err != nil {
			return err
		}
		if err := txUser.IncrementProfileCounter(ctx, comment.UserId, "comment_count"); err != nil {
			return err
		}
		if comment.RootCommentId != 0 {
			if err := txComment.IncrementReplyCount(ctx, comment.RootCommentId); err != nil {
				return err
			}
		}
		if comment.ParentCommentId != 0 && comment.ParentCommentId != comment.RootCommentId {
			if err := txComment.IncrementReplyCount(ctx, comment.ParentCommentId); err != nil {
				return err
			}
		}
		if comment.TargetType == constant.TargetPost {
			return txPost.IncrementCommentCount(ctx, comment.TargetId)
		}
		if comment.TargetType == constant.TargetVideo {
			return txVideo.IncrementCommentCount(ctx, comment.TargetId)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return cs.toCommentDetail(ctx, comment, 0), nil
}

func (cs *CommentService) DeleteCommentById(ctx context.Context, commentId, userId uint) error {
	comment, err := cs.commentRepo.FindById(ctx, commentId)
	if err != nil {
		return err
	}
	if comment.UserId != userId {
		return errors.New("you can't delete others' comment")
	}

	return cs.commentRepo.WithTransaction(ctx, func(tx *gorm.DB) error {
		txComment := cs.commentRepo.WithTx(tx)
		txUser := cs.userRepo.WithTx(tx)
		txPost := cs.postRepo.WithTx(tx)
		txVideo := cs.videoRepo.WithTx(tx)

		if comment.RootCommentId == 0 {
			delCnt, _ := txComment.DeleteByRootCommentId(ctx, commentId)
			if comment.TargetType == constant.TargetPost {
				txPost.DecrementCommentCount(ctx, comment.TargetId, int(delCnt))
			}
			if comment.TargetType == constant.TargetVideo {
				txVideo.DecrementCommentCount(ctx, comment.TargetId, int(delCnt))
			}
			txComment.DeleteById(ctx, commentId)
			return txUser.DecrementProfileCounter(ctx, comment.UserId, "comment_count")
		}

		delCnt, _ := txComment.DeleteByParentCommentId(ctx, commentId)
		if comment.TargetType == constant.TargetPost {
			txPost.DecrementCommentCount(ctx, comment.TargetId, int(delCnt))
		}
		if comment.TargetType == constant.TargetVideo {
			txVideo.DecrementCommentCount(ctx, comment.TargetId, int(delCnt))
		}
		txComment.DecrementReplyCount(ctx, comment.RootCommentId, int(delCnt))
		if comment.ParentCommentId != 0 && comment.ParentCommentId != comment.RootCommentId {
			txComment.DecrementReplyCount(ctx, comment.ParentCommentId, int(delCnt))
		}
		txComment.DeleteById(ctx, commentId)
		return txUser.DecrementProfileCounter(ctx, comment.UserId, "comment_count")
	})
}

func (cs *CommentService) ToggleLike(ctx context.Context, commentId, userId uint) error {
	return cs.commentRepo.WithTransaction(ctx, func(tx *gorm.DB) error {
		txAction := cs.actionRepo.WithTx(tx)
		txComment := cs.commentRepo.WithTx(tx)

		exists, _ := txAction.Exists(ctx, userId, commentId, constant.ActionLike, constant.TargetComment)
		if exists {
			txAction.Delete(ctx, userId, commentId, constant.ActionLike, constant.TargetComment)
			return txComment.DecrementLikeCount(ctx, commentId)
		}

		removed, _ := txAction.Delete(ctx, userId, commentId, constant.ActionDislike, constant.TargetComment)
		if removed > 0 {
			txComment.DecrementDislikeCount(ctx, commentId)
		}

		txAction.Create(ctx, &entity.Action{
			UserId:   userId,
			TargetId: commentId,
			ActType:  constant.ActionLike,
			TgtType:  constant.TargetComment,
		})
		return txComment.IncrementLikeCount(ctx, commentId)
	})
}

func (cs *CommentService) ToggleDislike(ctx context.Context, commentId, userId uint) error {
	return cs.commentRepo.WithTransaction(ctx, func(tx *gorm.DB) error {
		txAction := cs.actionRepo.WithTx(tx)
		txComment := cs.commentRepo.WithTx(tx)

		exists, _ := txAction.Exists(ctx, userId, commentId, constant.ActionDislike, constant.TargetComment)
		if exists {
			txAction.Delete(ctx, userId, commentId, constant.ActionDislike, constant.TargetComment)
			return txComment.DecrementDislikeCount(ctx, commentId)
		}

		removed, _ := txAction.Delete(ctx, userId, commentId, constant.ActionLike, constant.TargetComment)
		if removed > 0 {
			txComment.DecrementLikeCount(ctx, commentId)
		}

		txAction.Create(ctx, &entity.Action{
			UserId:   userId,
			TargetId: commentId,
			ActType:  constant.ActionDislike,
			TgtType:  constant.TargetComment,
		})
		return txComment.IncrementDislikeCount(ctx, commentId)
	})
}

func (cs *CommentService) ListCommentsByPostId(ctx context.Context, postId uint, page, pageSize int, viewerId uint) (*response.CommentList, error) {
	forbidden, _ := cs.postRepo.GetForbidComment(ctx, postId)
	if forbidden {
		return nil, errors.New("this post is forbidden to comment")
	}
	return cs.listCommentsByTarget(ctx, constant.TargetPost, postId, page, pageSize, viewerId)
}

func (cs *CommentService) ListCommentsByVideoId(ctx context.Context, videoId uint, page, pageSize int, viewerId uint) (*response.CommentList, error) {
	forbidden, _ := cs.videoRepo.GetForbidComment(ctx, videoId)
	if forbidden {
		return nil, errors.New("this video is forbidden to comment")
	}
	return cs.listCommentsByTarget(ctx, constant.TargetVideo, videoId, page, pageSize, viewerId)
}

func (cs *CommentService) ListCommentsByUserId(ctx context.Context, userId uint, page, pageSize int, viewerId uint) (*response.CommentList, error) {
	pagination, err := cs.commentRepo.ListCommentsByUserId(ctx, userId, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]*response.Comment, len(pagination.Items))
	for i, c := range pagination.Items {
		items[i] = cs.toCommentDetail(ctx, &c, viewerId)
	}
	return &response.CommentList{
		Items:    items,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Total:    pagination.Total,
	}, nil
}

// ======================== Internal ========================

func (cs *CommentService) listCommentsByTarget(ctx context.Context, targetType constant.TargetType, targetId uint, page, pageSize int, viewerId uint) (*response.CommentList, error) {
	pagination, err := cs.commentRepo.ListRootComments(ctx, targetType, targetId, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]*response.Comment, len(pagination.Items))
	for i, root := range pagination.Items {
		r := cs.toCommentDetail(ctx, &root, viewerId)
		r.ReplyComments = cs.getRepliesRecursive(ctx, r.CommentId, r.CommentId, viewerId)
		items[i] = r
	}
	return &response.CommentList{
		Items:    items,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Total:    pagination.Total,
	}, nil
}

func (cs *CommentService) getRepliesRecursive(ctx context.Context, rootId, parentId uint, viewerId uint) []*response.Comment {
	children, err := cs.commentRepo.FindRepliesByRootId(ctx, rootId, parentId)
	if err != nil {
		return nil
	}
	var res []*response.Comment
	for _, child := range children {
		rep := cs.toCommentDetail(ctx, &child, viewerId)
		rep.ReplyComments = cs.getRepliesRecursive(ctx, rootId, rep.CommentId, viewerId)
		res = append(res, rep)
	}
	return res
}

func (cs *CommentService) toCommentDetail(ctx context.Context, comment *entity.Comment, viewerId uint) *response.Comment {
	author, _ := cs.userService.GetUserById(ctx, comment.UserId, viewerId)
	var authorInfo response.UserInfo
	if author != nil {
		authorInfo = *author
	}

	isLiked := false
	isDisliked := false
	if viewerId > 0 {
		if exists, _ := cs.actionRepo.Exists(ctx, viewerId, comment.ID, constant.ActionLike, constant.TargetComment); exists {
			isLiked = true
		}
		if exists, _ := cs.actionRepo.Exists(ctx, viewerId, comment.ID, constant.ActionDislike, constant.TargetComment); exists {
			isDisliked = true
		}
	}

	return &response.Comment{
		Author:          authorInfo,
		CommentId:       comment.ID,
		Content:         comment.Content,
		CreatedAt:       comment.CreatedAt,
		UpdatedAt:       comment.UpdatedAt,
		DislikeCount:    comment.DislikeCount,
		IsDisliked:      isDisliked,
		IsLiked:         isLiked,
		LikeCount:       comment.LikeCount,
		ReplyCount:      comment.ReplyCount,
		ParentCommentId: comment.ParentCommentId,
		TargetId:        comment.TargetId,
		TargetType:      string(comment.TargetType),
		RootCommentId:   comment.RootCommentId,
		Env:             comment.EnvInfo,
	}
}
