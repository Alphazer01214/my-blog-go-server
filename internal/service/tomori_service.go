package service

import (
	"context"
	"errors"

	"blog.alphazer01214.top/internal/constant"
	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/response"
	"blog.alphazer01214.top/internal/utils"
	"gorm.io/gorm"
)

// TomoriService 超级管理员（TakamatsuTomori）论坛管理服务
// 所有方法均为管理员专用，跳过常规业务的所有权校验
type TomoriService struct{}

// ======================== 统计面板 ========================

// GetStats 获取论坛基础统计数据
func (ts *TomoriService) GetStats() (*response.TomoriStats, error) {
	stats := &response.TomoriStats{}
	db := global.GetDB()
	db.Model(&entity.User{}).Count(&stats.UserCount)
	db.Model(&entity.Post{}).Count(&stats.PostCount)
	db.Model(&entity.Comment{}).Count(&stats.CommentCount)
	db.Model(&entity.Video{}).Where("type = ?", "file").Count(&stats.FileCount)
	db.Model(&entity.Video{}).Where("type = ?", "video").Count(&stats.VideoCount)
	db.Model(&entity.Agent{}).Count(&stats.AgentCount)
	db.Model(&entity.ChatSession{}).Count(&stats.ChatCount)
	return stats, nil
}

// ======================== 用户管理 ========================

// ListUsers 分页查询所有用户（含 role、banned 状态，不受隐私设置限制）
func (ts *TomoriService) ListUsers(page, pageSize int) (*response.TomoriUserList, error) {
	var users []entity.User
	var total int64
	db := global.GetDB().Model(&entity.User{})
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}
	offset := (page - 1) * pageSize
	if err := db.Order("id desc").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, err
	}
	items := make([]response.UserInfo, len(users))
	for i, u := range users {
		items[i] = Service.UserService.toUserInfo(&u, u.ID)
	}
	return &response.TomoriUserList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

// BanUser 封禁或解封用户
func (ts *TomoriService) BanUser(userId uint, banned bool) error {
	result := global.GetDB().Model(&entity.User{}).Where("id = ?", userId).Update("banned", banned)
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return result.Error
}

// SetRole 修改用户角色
func (ts *TomoriService) SetRole(userId uint, role entity.RoleType) error {
	result := global.GetDB().Model(&entity.User{}).Where("id = ?", userId).Update("role", role)
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return result.Error
}

// ResetPassword 重置用户密码（管理员强制修改）
func (ts *TomoriService) ResetPassword(userId uint, newPassword string) error {
	if !utils.IsPasswordValid(newPassword) {
		return errors.New("password must be 6-256 characters")
	}
	hashed := utils.EncryptPassword(newPassword)
	result := global.GetDB().Model(&entity.User{}).Where("id = ?", userId).Update("password", hashed)
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return result.Error
}

// ForceDeleteUser 强制删除用户：级联删除 profile、setting、关注关系
func (ts *TomoriService) ForceDeleteUser(userId uint) error {
	return global.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userId).Delete(&entity.UserProfile{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userId).Delete(&entity.UserSetting{}).Error; err != nil {
			return err
		}
		if err := tx.Where("follower_id = ?", userId).Or("following_id = ?", userId).Delete(&entity.UserFollow{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", userId).Delete(&entity.User{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// ======================== 帖子管理 ========================

// ListAllPosts 分页查询所有帖子（含私密，无权限过滤）
func (ts *TomoriService) ListAllPosts(page, pageSize int) (*response.PostList, error) {
	var posts []entity.Post
	var total int64
	db := global.GetDB().Model(&entity.Post{})
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}
	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&posts).Error; err != nil {
		return nil, err
	}
	items := make([]response.PostDetail, len(posts))
	for i, post := range posts {
		author, err := Service.UserService.GetUserInfoById(post.UserId, 0)
		if err != nil {
			author = response.UserInfo{}
		}
		items[i] = *Service.PostService.toPostDetail(&post, &author, 0)
	}
	return &response.PostList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

// ForceDeletePost 管理员删除任意帖子，跳过作者校验
func (ts *TomoriService) ForceDeletePost(id uint) error {
	return Service.PostService.DeleteById(id)
}

// ======================== 评论管理 ========================

// ListAllComments 分页查询所有评论（平铺，无树结构）
func (ts *TomoriService) ListAllComments(page, pageSize int) (*response.CommentList, error) {
	var comments []entity.Comment
	var total int64
	db := global.GetDB().Model(&entity.Comment{})
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}
	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&comments).Error; err != nil {
		return nil, err
	}
	items := make([]*response.Comment, len(comments))
	for i, c := range comments {
		items[i] = Service.CommentService.toResponseComment(&c, 0)
	}
	return &response.CommentList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

// ForceDeleteComment 管理员删除任意评论（含级联子评论），跳过作者校验
func (ts *TomoriService) ForceDeleteComment(commentId uint) error {
	comment, err := Service.CommentService.getEntityByCommentId(commentId)
	if err != nil {
		return err
	}
	return global.GetDB().Transaction(func(tx *gorm.DB) error {
		if comment.RootCommentId == 0 {
			// 根评论：级联删除所有子回复
			res := tx.Where("root_comment_id = ?", commentId).Delete(&entity.Comment{})
			delCnt := res.RowsAffected
			if comment.TargetType == constant.TargetPost {
				if err := tx.Model(&entity.Post{}).Where("id = ?", comment.TargetId).
					Update("comment_count", gorm.Expr("GREATEST(comment_count - ?, 0)", delCnt)).Error; err != nil {
					return err
				}
			}
			if comment.TargetType == constant.TargetVideo {
				if err := tx.Model(&entity.Video{}).Where("id = ?", comment.TargetId).
					Update("comment_count", gorm.Expr("GREATEST(comment_count - ?, 0)", delCnt)).Error; err != nil {
					return err
				}
			}
			if err := tx.Delete(&entity.Comment{}, commentId).Error; err != nil {
				return err
			}
			return tx.Model(&entity.UserProfile{}).Where("user_id = ?", comment.UserId).
				Update("comment_count", gorm.Expr("GREATEST(comment_count - 1, 0)")).Error
		}
		// 子回复：级联删除以该评论为父评论的所有子回复
		res := tx.Where("parent_comment_id = ?", commentId).Delete(&entity.Comment{})
		delCnt := res.RowsAffected
		if comment.TargetType == constant.TargetPost {
			if err := tx.Model(&entity.Post{}).Where("id = ?", comment.TargetId).
				Update("comment_count", gorm.Expr("GREATEST(comment_count - ?, 0)", delCnt)).Error; err != nil {
				return err
			}
		}
		if comment.TargetType == constant.TargetVideo {
			if err := tx.Model(&entity.Video{}).Where("id = ?", comment.TargetId).
				Update("comment_count", gorm.Expr("GREATEST(comment_count - ?, 0)", delCnt)).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&entity.Comment{}).Where("id = ?", comment.RootCommentId).
			Update("reply_count", gorm.Expr("GREATEST(reply_count - ?, 0)", delCnt)).Error; err != nil {
			return err
		}
		if comment.ParentCommentId != 0 && comment.ParentCommentId != comment.RootCommentId {
			if err := tx.Model(&entity.Comment{}).Where("id = ?", comment.ParentCommentId).
				Update("reply_count", gorm.Expr("GREATEST(reply_count - ?, 0)", delCnt)).Error; err != nil {
				return err
			}
		}
		if err := tx.Delete(&entity.Comment{}, commentId).Error; err != nil {
			return err
		}
		return tx.Model(&entity.UserProfile{}).Where("user_id = ?", comment.UserId).
			Update("comment_count", gorm.Expr("GREATEST(comment_count - 1, 0)")).Error
	})
}

// ======================== 文件/视频管理 ========================

// ListAllFiles 分页查询所有文件（含私密）
func (ts *TomoriService) ListAllFiles(page, pageSize int) (*response.FileList, error) {
	var videos []entity.Video
	var total int64
	db := global.GetDB().Model(&entity.Video{}).Where("type = ?", "file")
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}
	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&videos).Error; err != nil {
		return nil, err
	}
	items := make([]response.FileDetail, len(videos))
	for i, v := range videos {
		items[i] = *Service.FileService.toFileDetail(&v)
	}
	return &response.FileList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

// ForceDeleteFile 管理员删除任意文件，跳过作者校验
func (ts *TomoriService) ForceDeleteFile(id uint) error {
	var video entity.Video
	if err := global.GetDB().Where("id = ?", id).First(&video).Error; err != nil {
		return err
	}
	return Service.FileService.DeleteFile(id, video.UserId)
}

// ======================== 视频管理 ========================

// ListAllVideos 分页查询所有视频（含私密）
func (ts *TomoriService) ListAllVideos(page, pageSize int) (*response.VideoList, error) {
	var videos []entity.Video
	var total int64
	db := global.GetDB().Model(&entity.Video{}).Where("type = ?", "video")
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}
	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&videos).Error; err != nil {
		return nil, err
	}
	items := make([]response.VideoDetail, len(videos))
	for i, v := range videos {
		author, err := Service.UserService.GetUserInfoById(v.UserId, 0)
		if err != nil {
			author = response.UserInfo{}
		}
		items[i] = *Service.VideoService.toVideoDetail(&v, &author, 0)
	}
	return &response.VideoList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

// ForceDeleteVideo 管理员删除任意视频，跳过作者校验
func (ts *TomoriService) ForceDeleteVideo(id uint) error {
	return Service.VideoService.DeleteById(id)
}

// ======================== AI 智能体管理 ========================

// ListAllAgents 查询所有用户创建的智能体
func (ts *TomoriService) ListAllAgents() (*response.TomoriAgentList, error) {
	var agents []entity.Agent
	var total int64
	db := global.GetDB().Model(&entity.Agent{})
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}
	if err := db.Order("id desc").Find(&agents).Error; err != nil {
		return nil, err
	}
	items := make([]response.TomoriAgentInfo, len(agents))
	for i, a := range agents {
		var user entity.User
		username := ""
		if err := global.GetDB().Where("id = ?", a.UserId).First(&user).Error; err == nil {
			username = user.Username
		}
		items[i] = response.TomoriAgentInfo{
			AgentId:   a.ID,
			UserId:    a.UserId,
			Username:  username,
			Name:      a.Name,
			ModelName: a.ModelName,
			Provider:  a.Provider,
			Activate:  a.Activate,
			CreatedAt: a.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return &response.TomoriAgentList{Items: items, Total: total}, nil
}

// ForceDeleteAgent 管理员删除任意智能体，跳过所有者校验
func (ts *TomoriService) ForceDeleteAgent(agentId uint) error {
	result := global.GetDB().Delete(&entity.Agent{}, agentId)
	if result.RowsAffected == 0 {
		return errors.New("agent not found")
	}
	return result.Error
}

// ======================== 聊天会话管理 ========================

// ListAllChats 分页查询所有聊天会话
func (ts *TomoriService) ListAllChats(page, pageSize int) ([]entity.ChatSession, int64, error) {
	var sessions []entity.ChatSession
	var total int64
	db := global.GetDB().Model(&entity.ChatSession{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := db.Order("updated_at desc").Offset(offset).Limit(pageSize).Find(&sessions).Error; err != nil {
		return nil, 0, err
	}
	return sessions, total, nil
}

// ForceDeleteChat 管理员删除任意聊天会话，跳过所有者校验
func (ts *TomoriService) ForceDeleteChat(chatUuid string) error {
	result := global.GetDB().Where("uuid = ?", chatUuid).Delete(&entity.ChatSession{})
	if result.RowsAffected == 0 {
		return errors.New("chat session not found")
	}
	return result.Error
}

// ======================== 系统维护 ========================

// ListBlacklist 分页查询 Token 黑名单（从 Redis）
func (ts *TomoriService) ListBlacklist(page, pageSize int) (*response.TomoriBlacklist, error) {
	ctx := context.Background()
	var allKeys []string
	var cursor uint64
	for {
		keys, nextCursor, err := global.GetRedis().Scan(ctx, cursor, "token:blacklist:*", 200).Result()
		if err != nil {
			return nil, err
		}
		allKeys = append(allKeys, keys...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	total := int64(len(allKeys))
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > len(allKeys) {
		start = len(allKeys)
	}
	if end > len(allKeys) {
		end = len(allKeys)
	}
	pageKeys := allKeys[start:end]

	items := make([]response.TomoriBlacklistItem, len(pageKeys))
	for i, key := range pageKeys {
		token := key[len("token:blacklist:"):]
		ttl, _ := global.GetRedis().TTL(ctx, key).Result()
		tokenPreview := token
		if len(token) > 20 {
			tokenPreview = token[:20] + "..."
		}
		items[i] = response.TomoriBlacklistItem{
			ID:        key,
			Token:     tokenPreview,
			CreatedAt: ttl.String(),
		}
	}
	return &response.TomoriBlacklist{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

// ClearBlacklist 清空 Token 黑名单（从 Redis）
func (ts *TomoriService) ClearBlacklist() error {
	ctx := context.Background()
	var cursor uint64
	for {
		keys, nextCursor, err := global.GetRedis().Scan(ctx, cursor, "token:blacklist:*", 200).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := global.GetRedis().Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}
