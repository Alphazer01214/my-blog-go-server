package service

import (
	"context"
	"encoding/json"
	"time"

	"blog.alphazer01214.top/internal/constant"
	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	"blog.alphazer01214.top/internal/search"
	pkgKafka "blog.alphazer01214.top/pkg/kafka"
	"blog.alphazer01214.top/pkg/cache"

	"gorm.io/gorm"
)

type PostService struct{}

// postCache 帖子详情缓存（5 分钟 TTL）
var postCache *cache.RedisCache[entity.Post]

// SearchSvc Elasticsearch 搜索服务（由 main.go 初始化注入）
var SearchSvc *search.SearchService

// getPostCache 懒初始化帖子缓存
func getPostCache() *cache.RedisCache[entity.Post] {
	if postCache == nil {
		if rdb := global.GetRedis(); rdb != nil {
			postCache = cache.NewRedisCache[entity.Post](rdb, "cache:post:", 5*time.Minute)
		}
	}
	return postCache
}

func (ps *PostService) Create(post *entity.Post) (*response.PostDetail, error) {
	if err := ps.create(post); err != nil {
		return nil, err
	}
	author, err := Service.UserService.GetUserInfoById(post.UserId, 0)
	if err != nil {
		return nil, err
	}
	if err := global.GetDB().Model(&entity.UserProfile{}).Where("user_id = ?", post.UserId).Update("post_count", gorm.Expr("post_count + 1")).Error; err != nil {
		return nil, err
	}

	// 发布 Kafka 事件：帖子创建
	publishPostEvent(pkgKafka.ActionCreate, post.UserId, post.ID)

	return ps.toPostDetail(post, &author, 0), nil
}

func (ps *PostService) GetPostByPostId(id uint, viewerId uint) (*response.PostDetail, error) {
	post, err := ps.getPostEntityById(id)
	ps.visit(id, viewerId)
	if err != nil {
		return nil, err
	}
	author, err := Service.UserService.GetUserInfoById(post.UserId, 0)
	if err != nil {
		return nil, err
	}
	viewer, err := Service.UserService.GetUserInfoById(viewerId, 0)
	if err != nil {
		return nil, err
	}
	if !post.Public && post.UserId != viewerId && viewer.Role != entity.RoleTakamatsuTomori {
		post.Content = "this is a private post"
		//return ps.toPostDetail(post, &author, viewerId), nil
	}
	return ps.toPostDetail(post, &author, viewerId), nil
}

func (ps *PostService) GetAllPosts(page, pageSize int, viewerId uint) (response.PostList, error) {
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
		viewer, err := Service.UserService.GetUserInfoById(viewerId, 0)
		if err != nil {
			return response.PostList{}, err
		}
		if !post.Public && viewerId != author.UserId && viewer.Role != entity.RoleTakamatsuTomori {
			continue
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

func (ps *PostService) SearchPost(req *request.PostSearchRequest, page, pageSize int, viewerId uint) (response.PostList, error) {
	// 优先使用 Elasticsearch
	if SearchSvc != nil && SearchSvc.IsAvailable(context.Background()) {
		return ps.searchWithES(req, page, pageSize, viewerId)
	}
	// 降级到 PostgreSQL ILIKE
	return ps.searchWithPG(req, page, pageSize, viewerId)
}

// searchWithES 使用 Elasticsearch 搜索
func (ps *PostService) searchWithES(req *request.PostSearchRequest, page, pageSize int, viewerId uint) (response.PostList, error) {
	esReq := &search.SearchRequest{
		Keyword:  req.Keyword,
		Page:     page,
		PageSize: pageSize,
	}

	result, err := SearchSvc.SearchPosts(context.Background(), esReq)
	if err != nil {
		// ES 失败，降级到 PG
		return ps.searchWithPG(req, page, pageSize, viewerId)
	}

	return SearchSvc.ConvertToPostList(result, func(hit *search.PostSearchHit) response.PostDetail {
		return response.PostDetail{
			ID:           hit.ID,
			Title:        hit.Title,
			Content:      hit.Content,
			Tags:         nil, // ES 返回的是 []string，需要转换
			Category:     hit.Category,
			UserId:       hit.UserId,
			ViewCount:    hit.ViewCount,
			LikeCount:    hit.LikeCount,
			CommentCount: hit.CommentCount,
			CreatedAt:    hit.CreatedAt,
		}
	}), nil
}

// searchWithPG 使用 PostgreSQL ILIKE 搜索（降级方案）
func (ps *PostService) searchWithPG(req *request.PostSearchRequest, page, pageSize int, viewerId uint) (response.PostList, error) {
	var entities []entity.Post
	var total int64
	qKeyword := "%" + req.Keyword + "%"

	db := global.GetDB().Model(&entity.Post{}).Where("title ILIKE ? OR content ILIKE ?", qKeyword, qKeyword)

	if err := db.Count(&total).Error; err != nil {
		return response.PostList{}, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&entities).Error; err != nil {
		return response.PostList{}, err
	}

	items := make([]response.PostDetail, 0, len(entities))
	for _, et := range entities {
		userinfo, err := Service.UserService.GetUserInfoById(et.UserId, viewerId)
		if err != nil {
			continue
		}
		pd := ps.toPostDetail(&et, &userinfo, viewerId)
		items = append(items, *pd)
	}

	return response.PostList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (ps *PostService) DeleteById(id uint) error {
	post, err := ps.getPostEntityById(id)
	if err != nil {
		return err
	}
	if err := global.GetDB().Delete(&entity.Post{}, id).Error; err != nil {
		return err
	}

	// 发布 Kafka 事件：帖子删除 + 缓存失效
	publishPostEvent(pkgKafka.ActionDelete, post.UserId, id)
	publishCacheInvalidation("post", id)

	return global.GetDB().Model(&entity.UserProfile{}).Where("user_id = ?", post.UserId).Update("post_count", gorm.Expr("GREATEST(post_count - 1, 0)")).Error
}

func (ps *PostService) Update(id uint, req request.PostUpdateRequest) error {
	dbPost, err := ps.getPostEntityById(id)
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

	if err := global.GetDB().Save(dbPost).Error; err != nil {
		return err
	}

	// 发布缓存失效事件
	publishCacheInvalidation("post", id)
	publishPostEvent(pkgKafka.ActionUpdate, dbPost.UserId, id)

	return nil
}

func (ps *PostService) Like(postId, userId uint) error {
	err := global.GetDB().Transaction(func(tx *gorm.DB) error {
		var existing entity.Action
		// 如果 like 了就取消
		result := tx.Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			postId, userId, constant.TargetPost, constant.ActionLike).First(&existing)
		if result.RowsAffected > 0 {
			if err := tx.Delete(&existing).Error; err != nil {
				return err
			}
			// 取消点赞 → 发送通知事件 + 缓存失效
			go publishPostNotification(pkgKafka.ActionUnlike, userId, postId, "取消了点赞")
			go publishCacheInvalidation("post", postId)
			return tx.Model(&entity.Post{}).Where("id = ?", postId).
				Update("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error
		}

		// 原本是 dislike， 取消 disklike 并 like
		r := tx.Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			postId, userId, constant.TargetPost, constant.ActionDislike).Delete(&entity.Action{})
		if r.RowsAffected > 0 {
			if err := tx.Model(&entity.Post{}).Where("id = ?", postId).
				Update("dislike_count", gorm.Expr("GREATEST(dislike_count - 1, 0)")).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(&entity.Action{
			UserId:   userId,
			TargetId: postId,
			ActType:  constant.ActionLike,
			TgtType:  constant.TargetPost,
		}).Error; err != nil {
			return err
		}

		// 点赞成功 → 发送通知事件 + 缓存失效
		go publishPostNotification(pkgKafka.ActionLike, userId, postId, "赞了你的帖子")
		go publishCacheInvalidation("post", postId)

		return tx.Model(&entity.Post{}).Where("id = ?", postId).
			Update("like_count", gorm.Expr("like_count + 1")).Error
	})
	return err
}

func (ps *PostService) Dislike(postId, userId uint) error {
	return global.GetDB().Transaction(func(tx *gorm.DB) error {
		var existing entity.Action
		result := tx.Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			postId, userId, constant.TargetPost, constant.ActionDislike).First(&existing)
		if result.RowsAffected > 0 {
			if err := tx.Delete(&existing).Error; err != nil {
				return err
			}
			return tx.Model(&entity.Post{}).Where("id = ?", postId).
				Update("dislike_count", gorm.Expr("GREATEST(dislike_count - 1, 0)")).Error
		}

		r := tx.Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			postId, userId, constant.TargetPost, constant.ActionLike).Delete(&entity.Action{})
		if r.RowsAffected > 0 {
			if err := tx.Model(&entity.Post{}).Where("id = ?", postId).
				Update("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(&entity.Action{
			UserId:   userId,
			TargetId: postId,
			ActType:  constant.ActionDislike,
			TgtType:  constant.TargetPost,
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
			postId, userId, constant.TargetPost, constant.ActionFavorite).First(&existing)
		if result.RowsAffected > 0 {
			if err := tx.Delete(&existing).Error; err != nil {
				return err
			}
			return tx.Model(&entity.Post{}).Where("id = ?", postId).
				Update("favorite_count", gorm.Expr("GREATEST(favorite_count - 1, 0)")).Error
		}

		if err := tx.Create(&entity.Action{
			UserId:   userId,
			TargetId: postId,
			ActType:  constant.ActionFavorite,
			TgtType:  constant.TargetPost,
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
			UserId:    userId,
			TargetId:  postId,
			ActType:   constant.ActionShare,
			TgtType:   constant.TargetPost,
			ExtraInfo: extraJSON,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&entity.Post{}).Where("id = ?", postId).
			Update("share_count", gorm.Expr("share_count + 1")).Error
	})
}

func (ps *PostService) GetPostsByUserId(userId uint, page, pageSize int) (response.PostList, error) {
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

func (ps *PostService) getPostEntityById(id uint) (*entity.Post, error) {
	pc := getPostCache()
	if pc != nil {
		return pc.GetOrLoad(context.Background(), cache.CacheKey(id), func() (*entity.Post, error) {
			var post entity.Post
			err := global.GetDB().Where("id = ?", id).First(&post).Error
			if err != nil {
				return nil, err
			}
			return &post, nil
		})
	}
	// 缓存不可用，直接查 DB
	var post entity.Post
	err := global.GetDB().Where("id = ?", id).First(&post).Error
	return &post, err
}

func (ps *PostService) GetPostEntityById(id uint) (*entity.Post, error) {
	return ps.getPostEntityById(id)
}

func (ps *PostService) visit(postId uint, viewerId uint) {
	global.GetDB().Model(&entity.Post{}).Where("id = ?", postId).
		Update("view_count", gorm.Expr("view_count + 1"))
}

func (ps *PostService) isForbidComment(postId uint, viewerId uint) bool {
	var res bool
	err := global.GetDB().Model(&entity.Post{}).Where("ID = ?", postId).Pluck("forbid_comment", &res).Error
	return err == nil && res
}

func (ps *PostService) toPostDetail(post *entity.Post, author *response.UserInfo, viewerId uint) *response.PostDetail {
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
		var like entity.Action
		if global.GetDB().Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			post.ID, viewerId, constant.TargetPost, constant.ActionLike).First(&like).Error == nil {
			pd.IsLiked = true
		}
		var dislike entity.Action
		if global.GetDB().Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			post.ID, viewerId, constant.TargetPost, constant.ActionDislike).First(&dislike).Error == nil {
			pd.IsDisliked = true
		}
		var fav entity.Action
		if global.GetDB().Where("target_id = ? AND user_id = ? AND target_type = ? AND action_type = ?",
			post.ID, viewerId, constant.TargetPost, constant.ActionFavorite).First(&fav).Error == nil {
			pd.IsFavorited = true
		}
	}

	return pd
}

// ConvertToDetail 公开版本，供外部包调用
func (ps *PostService) ConvertToDetail(post *entity.Post, author *response.UserInfo, viewerId uint) *response.PostDetail {
	return ps.toPostDetail(post, author, viewerId)
}

func (ps *PostService) GetFavoritesByUserId(userId uint, page, pageSize int) (response.PostList, error) {
	var actions []entity.Action
	var total int64

	db := global.GetDB().Model(&entity.Action{}).Where(
		"user_id = ? AND target_type = ? AND action_type = ?",
		userId, constant.TargetPost, constant.ActionFavorite,
	)

	if err := db.Count(&total).Error; err != nil {
		return response.PostList{}, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&actions).Error; err != nil {
		return response.PostList{}, err
	}

	var postIds []uint
	for _, action := range actions {
		postIds = append(postIds, action.TargetId)
	}

	var posts []entity.Post
	if err := global.GetDB().Where("id IN ?", postIds).Find(&posts).Error; err != nil {
		return response.PostList{}, err
	}

	postMap := make(map[uint]entity.Post)
	for _, post := range posts {
		postMap[post.ID] = post
	}

	items := make([]response.PostDetail, 0, len(actions))
	for _, action := range actions {
		if post, ok := postMap[action.TargetId]; ok {
			author, err := Service.UserService.GetUserInfoById(post.UserId, 0)
			if err != nil {
				continue
			}
			pd := ps.toPostDetail(&post, &author, userId)
			items = append(items, *pd)
		}
	}

	return response.PostList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}
