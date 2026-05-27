package service

import (
	"context"
	"encoding/json"

	"blog.alphazer01214.top/internal/constant"
	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/repository"
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"

	"gorm.io/gorm"
)

type PostService struct {
	postRepo    repository.PostRepository
	userRepo    repository.UserRepository
	actionRepo  repository.ActionRepository
	userService *UserService
}

func NewPostService(
	postRepo repository.PostRepository,
	userRepo repository.UserRepository,
	actionRepo repository.ActionRepository,
	userService *UserService,
) *PostService {
	return &PostService{
		postRepo:    postRepo,
		userRepo:    userRepo,
		actionRepo:  actionRepo,
		userService: userService,
	}
}

func (ps *PostService) Create(ctx context.Context, post *entity.Post) (*response.PostDetail, error) {
	if err := ps.postRepo.Create(ctx, post); err != nil {
		return nil, err
	}
	if err := ps.userRepo.IncrementProfileCounter(ctx, post.UserId, "post_count"); err != nil {
		return nil, err
	}
	author, err := ps.userService.GetUserById(ctx, post.UserId, 0)
	if err != nil {
		return nil, err
	}
	return ps.toPostDetail(ctx, post, author, 0), nil
}

func (ps *PostService) GetPostById(ctx context.Context, id uint, viewerId uint) (*response.PostDetail, error) {
	post, err := ps.postRepo.FindById(ctx, id)
	if err != nil {
		return nil, err
	}
	_ = ps.postRepo.IncrementViewCount(ctx, id)
	author, err := ps.userService.GetUserById(ctx, post.UserId, 0)
	if err != nil {
		return nil, err
	}
	return ps.toPostDetail(ctx, post, author, viewerId), nil
}

func (ps *PostService) ListPosts(ctx context.Context, page, pageSize int, viewerId uint) (*response.PostList, error) {
	pagination, err := ps.postRepo.ListPosts(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]response.PostDetail, len(pagination.Items))
	for i, post := range pagination.Items {
		author, _ := ps.userService.GetUserById(ctx, post.UserId, 0)
		items[i] = *ps.toPostDetail(ctx, &post, author, viewerId)
	}
	return &response.PostList{
		Items:    items,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Total:    pagination.Total,
	}, nil
}

func (ps *PostService) ListPostsByUserId(ctx context.Context, userId uint, page, pageSize int, viewerId uint) (*response.PostList, error) {
	pagination, err := ps.postRepo.ListPostsByUserId(ctx, userId, page, pageSize)
	if err != nil {
		return nil, err
	}
	author, _ := ps.userService.GetUserById(ctx, userId, 0)
	items := make([]response.PostDetail, len(pagination.Items))
	for i, post := range pagination.Items {
		items[i] = *ps.toPostDetail(ctx, &post, author, viewerId)
	}
	return &response.PostList{
		Items:    items,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Total:    pagination.Total,
	}, nil
}

func (ps *PostService) SearchPosts(ctx context.Context, keyword string, page, pageSize int, viewerId uint) (*response.PostList, error) {
	pagination, err := ps.postRepo.SearchByKeyword(ctx, keyword, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]response.PostDetail, len(pagination.Items))
	for i, post := range pagination.Items {
		author, _ := ps.userService.GetUserById(ctx, post.UserId, 0)
		items[i] = *ps.toPostDetail(ctx, &post, author, viewerId)
	}
	return &response.PostList{
		Items:    items,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Total:    pagination.Total,
	}, nil
}

func (ps *PostService) UpdatePost(ctx context.Context, id uint, req request.PostUpdateRequest) error {
	dbPost, err := ps.postRepo.FindById(ctx, id)
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
	dbPost.ForbidComment = req.ForbidComment
	dbPost.ForbidShare = req.ForbidShare
	dbPost.EnvInfo = req.Env
	return ps.postRepo.Save(ctx, dbPost)
}

func (ps *PostService) DeletePostById(ctx context.Context, id uint) error {
	post, err := ps.postRepo.FindById(ctx, id)
	if err != nil {
		return err
	}
	if err := ps.postRepo.DeleteById(ctx, id); err != nil {
		return err
	}
	return ps.userRepo.DecrementProfileCounter(ctx, post.UserId, "post_count")
}

// ======================== Actions ========================

func (ps *PostService) ToggleLike(ctx context.Context, postId, userId uint) error {
	return ps.postRepo.WithTransaction(ctx, func(tx *gorm.DB) error {
		txAction := ps.actionRepo.WithTx(tx)
		txPost := ps.postRepo.WithTx(tx)

		exists, _ := txAction.Exists(ctx, userId, postId, constant.ActionLike, constant.TargetPost)
		if exists {
			txAction.Delete(ctx, userId, postId, constant.ActionLike, constant.TargetPost)
			return txPost.DecrementLikeCount(ctx, postId)
		}

		// Remove opposing dislike
		removed, _ := txAction.Delete(ctx, userId, postId, constant.ActionDislike, constant.TargetPost)
		if removed > 0 {
			txPost.DecrementDislikeCount(ctx, postId)
		}

		txAction.Create(ctx, &entity.Action{
			UserId:   userId,
			TargetId: postId,
			ActType:  constant.ActionLike,
			TgtType:  constant.TargetPost,
		})
		return txPost.IncrementLikeCount(ctx, postId)
	})
}

func (ps *PostService) ToggleDislike(ctx context.Context, postId, userId uint) error {
	return ps.postRepo.WithTransaction(ctx, func(tx *gorm.DB) error {
		txAction := ps.actionRepo.WithTx(tx)
		txPost := ps.postRepo.WithTx(tx)

		exists, _ := txAction.Exists(ctx, userId, postId, constant.ActionDislike, constant.TargetPost)
		if exists {
			txAction.Delete(ctx, userId, postId, constant.ActionDislike, constant.TargetPost)
			return txPost.DecrementDislikeCount(ctx, postId)
		}

		removed, _ := txAction.Delete(ctx, userId, postId, constant.ActionLike, constant.TargetPost)
		if removed > 0 {
			txPost.DecrementLikeCount(ctx, postId)
		}

		txAction.Create(ctx, &entity.Action{
			UserId:   userId,
			TargetId: postId,
			ActType:  constant.ActionDislike,
			TgtType:  constant.TargetPost,
		})
		return txPost.IncrementDislikeCount(ctx, postId)
	})
}

func (ps *PostService) ToggleFavorite(ctx context.Context, postId, userId uint) error {
	return ps.postRepo.WithTransaction(ctx, func(tx *gorm.DB) error {
		txAction := ps.actionRepo.WithTx(tx)
		txPost := ps.postRepo.WithTx(tx)

		exists, _ := txAction.Exists(ctx, userId, postId, constant.ActionFavorite, constant.TargetPost)
		if exists {
			txAction.Delete(ctx, userId, postId, constant.ActionFavorite, constant.TargetPost)
			return txPost.DecrementFavoriteCount(ctx, postId)
		}

		txAction.Create(ctx, &entity.Action{
			UserId:   userId,
			TargetId: postId,
			ActType:  constant.ActionFavorite,
			TgtType:  constant.TargetPost,
		})
		return txPost.IncrementFavoriteCount(ctx, postId)
	})
}

func (ps *PostService) Share(ctx context.Context, postId, userId uint, shareInfo entity.ShareInfo) error {
	extraJSON, err := json.Marshal(shareInfo)
	if err != nil {
		return err
	}
	return ps.postRepo.WithTransaction(ctx, func(tx *gorm.DB) error {
		txAction := ps.actionRepo.WithTx(tx)
		txPost := ps.postRepo.WithTx(tx)

		if err := txAction.Create(ctx, &entity.Action{
			UserId:    userId,
			TargetId:  postId,
			ActType:   constant.ActionShare,
			TgtType:   constant.TargetPost,
			ExtraInfo: extraJSON,
		}); err != nil {
			return err
		}
		return txPost.IncrementShareCount(ctx, postId)
	})
}

// ======================== Internal ========================

func (ps *PostService) toPostDetail(ctx context.Context, post *entity.Post, author *response.UserInfo, viewerId uint) *response.PostDetail {
	pd := &response.PostDetail{
		ID:            post.ID,
		CreatedAt:     post.CreatedAt,
		UpdatedAt:     post.UpdatedAt,
		Title:         post.Title,
		Cover:         post.Cover,
		UserId:        post.UserId,
		Tags:          post.Tags,
		Category:      post.Category,
		Keywords:      post.Keywords,
		Content:       post.Content,
		ViewCount:     post.ViewCount,
		CommentCount:  post.CommentCount,
		LikeCount:     post.LikeCount,
		DislikeCount:  post.DislikeCount,
		FavoriteCount: post.FavoriteCount,
		ShareCount:    post.ShareCount,
		Env:           post.EnvInfo,
		Public:        post.Public,
		ForbidComment: post.ForbidComment,
		ForbidShare:   post.ForbidShare,
	}
	if author != nil {
		pd.Author = *author
	}

	if viewerId > 0 {
		if exists, _ := ps.actionRepo.Exists(ctx, viewerId, post.ID, constant.ActionLike, constant.TargetPost); exists {
			pd.IsLiked = true
		}
		if exists, _ := ps.actionRepo.Exists(ctx, viewerId, post.ID, constant.ActionDislike, constant.TargetPost); exists {
			pd.IsDisliked = true
		}
		if exists, _ := ps.actionRepo.Exists(ctx, viewerId, post.ID, constant.ActionFavorite, constant.TargetPost); exists {
			pd.IsFavorited = true
		}
	}

	return pd
}
