package service

import (
	"context"
	"errors"
	"fmt"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/repository"
	"blog.alphazer01214.top/internal/response"
	"blog.alphazer01214.top/internal/utils"

	"gorm.io/gorm"
)

type TomoriService struct {
	userRepo    repository.UserRepository
	postRepo    repository.PostRepository
	commentRepo repository.CommentRepository
	videoRepo   repository.VideoRepository
	aiRepo      repository.AIRepository
	userService *UserService
	postService *PostService
	fileService *FileService
}

func NewTomoriService(
	userRepo repository.UserRepository,
	postRepo repository.PostRepository,
	commentRepo repository.CommentRepository,
	videoRepo repository.VideoRepository,
	aiRepo repository.AIRepository,
	userService *UserService,
	postService *PostService,
	fileService *FileService,
) *TomoriService {
	return &TomoriService{
		userRepo:    userRepo,
		postRepo:    postRepo,
		commentRepo: commentRepo,
		videoRepo:   videoRepo,
		aiRepo:      aiRepo,
		userService: userService,
		postService: postService,
		fileService: fileService,
	}
}

// ======================== Stats ========================

func (ts *TomoriService) GetStats(ctx context.Context) (*response.TomoriStats, error) {
	stats := &response.TomoriStats{}
	db := ts.userRepo.(*repository.UserRepositoryImpl).DB
	db.WithContext(ctx).Model(&entity.User{}).Count(&stats.UserCount)
	db.WithContext(ctx).Model(&entity.Post{}).Count(&stats.PostCount)
	db.WithContext(ctx).Model(&entity.Comment{}).Count(&stats.CommentCount)
	db.WithContext(ctx).Model(&entity.Video{}).Count(&stats.FileCount)
	db.WithContext(ctx).Model(&entity.Agent{}).Count(&stats.AgentCount)
	db.WithContext(ctx).Model(&entity.ChatSession{}).Count(&stats.ChatCount)
	return stats, nil
}

// ======================== User Management ========================

func (ts *TomoriService) ListUsers(ctx context.Context, page, pageSize int) (*response.TomoriUserList, error) {
	pagination, err := ts.userRepo.ListUsers(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]response.UserInfo, len(pagination.Items))
	for i, u := range pagination.Items {
		info, _ := ts.userService.toUserInfo(ctx, &u, u.ID)
		items[i] = info
	}
	return &response.TomoriUserList{
		Items:    items,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Total:    pagination.Total,
	}, nil
}

func (ts *TomoriService) BanUser(ctx context.Context, userId uint, banned bool) error {
	affected, err := ts.userRepo.UpdateBanned(ctx, userId, banned)
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (ts *TomoriService) SetRole(ctx context.Context, userId uint, role entity.RoleType) error {
	affected, err := ts.userRepo.UpdateRole(ctx, userId, role)
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (ts *TomoriService) ResetPassword(ctx context.Context, userId uint, newPassword string) error {
	if !utils.IsPasswordValid(newPassword) {
		return errors.New("password must be 6-256 characters")
	}
	hashed := utils.EncryptPassword(newPassword)
	return ts.userRepo.UpdatePassword(ctx, userId, hashed)
}

func (ts *TomoriService) ForceDeleteUser(ctx context.Context, userId uint) error {
	return ts.userRepo.WithTransaction(ctx, func(tx *gorm.DB) error {
		txUser := ts.userRepo.WithTx(tx)
		if err := txUser.DeleteProfileByUserId(ctx, userId); err != nil {
			return err
		}
		if err := txUser.DeleteSettingByUserId(ctx, userId); err != nil {
			return err
		}
		if err := txUser.DeleteFollowsByUserId(ctx, userId); err != nil {
			return err
		}
		return txUser.DeleteById(ctx, userId)
	})
}

// ======================== Post Management ========================

func (ts *TomoriService) ListAllPosts(ctx context.Context, page, pageSize int) (*response.PostList, error) {
	pagination, err := ts.postRepo.ListPosts(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]response.PostDetail, len(pagination.Items))
	for i, post := range pagination.Items {
		author, _ := ts.userService.GetUserById(ctx, post.UserId, 0)
		items[i] = *ts.postService.toPostDetail(ctx, &post, author, 0)
	}
	return &response.PostList{
		Items:    items,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Total:    pagination.Total,
	}, nil
}

func (ts *TomoriService) ForceDeletePost(ctx context.Context, id uint) error {
	return ts.postService.DeletePostById(ctx, id)
}

// ======================== Comment Management ========================

func (ts *TomoriService) ListAllComments(ctx context.Context, page, pageSize int) (*response.CommentList, error) {
	pagination, err := ts.commentRepo.ListCommentsByUserId(ctx, 0, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]*response.Comment, len(pagination.Items))
	for i, c := range pagination.Items {
		items[i] = ts.commentService().toCommentDetail(ctx, &c, 0)
	}
	return &response.CommentList{
		Items:    items,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Total:    pagination.Total,
	}, nil
}

func (ts *TomoriService) ForceDeleteComment(ctx context.Context, commentId uint) error {
	comment, err := ts.commentRepo.FindById(ctx, commentId)
	if err != nil {
		return err
	}
	return ts.commentRepo.WithTransaction(ctx, func(tx *gorm.DB) error {
		txComment := ts.commentRepo.WithTx(tx)
		txUser := ts.userRepo.WithTx(tx)
		txPost := ts.postRepo.WithTx(tx)

		if comment.RootCommentId == 0 {
			delCnt, _ := txComment.DeleteByRootCommentId(ctx, commentId)
			txPost.DecrementCommentCount(ctx, comment.TargetId, int(delCnt))
			txComment.DeleteById(ctx, commentId)
			return txUser.DecrementProfileCounter(ctx, comment.UserId, "comment_count")
		}

		delCnt, _ := txComment.DeleteByParentCommentId(ctx, commentId)
		txPost.DecrementCommentCount(ctx, comment.TargetId, int(delCnt))
		txComment.DecrementReplyCount(ctx, comment.RootCommentId, int(delCnt))
		if comment.ParentCommentId != 0 && comment.ParentCommentId != comment.RootCommentId {
			txComment.DecrementReplyCount(ctx, comment.ParentCommentId, int(delCnt))
		}
		txComment.DeleteById(ctx, commentId)
		return txUser.DecrementProfileCounter(ctx, comment.UserId, "comment_count")
	})
}

// ======================== File/Video Management ========================

func (ts *TomoriService) ListAllFiles(ctx context.Context, page, pageSize int) (*response.FileList, error) {
	pagination, err := ts.videoRepo.ListVideos(ctx, false, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]response.FileDetail, len(pagination.Items))
	for i, v := range pagination.Items {
		items[i] = *ts.fileService.toFileDetail(&v)
	}
	return &response.FileList{
		Items:    items,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Total:    pagination.Total,
	}, nil
}

func (ts *TomoriService) ForceDeleteFile(ctx context.Context, id uint) error {
	return ts.fileService.DeleteFile(ctx, id, 0)
}

// ======================== Agent Management ========================

func (ts *TomoriService) ListAllAgents(ctx context.Context) (*response.TomoriAgentList, error) {
	var agents []entity.Agent
	db := ts.aiRepo.(*repository.AIRepositoryImpl).DB
	if err := db.WithContext(ctx).Model(&entity.Agent{}).Order("id desc").Find(&agents).Error; err != nil {
		return nil, err
	}
	items := make([]response.TomoriAgentInfo, len(agents))
	for i, a := range agents {
		username := ""
		if user, err := ts.userRepo.FindById(ctx, a.UserId); err == nil {
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
	return &response.TomoriAgentList{Items: items, Total: int64(len(items))}, nil
}

func (ts *TomoriService) ForceDeleteAgent(ctx context.Context, agentId uint) error {
	affected, err := ts.aiRepo.DeleteAgentById(ctx, agentId)
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("agent not found")
	}
	return nil
}

// ======================== Chat Session Management ========================

func (ts *TomoriService) ListAllChats(ctx context.Context, page, pageSize int) (*response.ChatSessionList, error) {
	db := ts.aiRepo.(*repository.AIRepositoryImpl).DB
	var sessions []entity.ChatSession
	var total int64
	q := db.WithContext(ctx).Model(&entity.ChatSession{})
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	offset := (page - 1) * pageSize
	if err := q.Order("updated_at desc").Offset(offset).Limit(pageSize).Find(&sessions).Error; err != nil {
		return nil, err
	}
	return &response.ChatSessionList{
		Items:    sessions,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (ts *TomoriService) ForceDeleteChat(ctx context.Context, chatUuid string) error {
	affected, err := ts.aiRepo.DeleteChatSessionByUuid(ctx, chatUuid)
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("chat session not found")
	}
	return nil
}

// ======================== Token Blacklist ========================

func (ts *TomoriService) ListBlacklist(ctx context.Context, page, pageSize int) (*response.TomoriBlacklist, error) {
	pagination, err := ts.userRepo.ListBlacklist(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]response.TomoriBlacklistItem, len(pagination.Items))
	for i, t := range pagination.Items {
		tokenPreview := t.Token
		if len(t.Token) > 20 {
			tokenPreview = t.Token[:20] + "..."
		}
		items[i] = response.TomoriBlacklistItem{
			ID:        fmt.Sprintf("%d", t.ID),
			Token:     tokenPreview,
			CreatedAt: t.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return &response.TomoriBlacklist{
		Items:    items,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Total:    pagination.Total,
	}, nil
}

func (ts *TomoriService) ClearBlacklist(ctx context.Context) error {
	return ts.userRepo.ClearBlacklist(ctx)
}

// ======================== Internal ========================

func (ts *TomoriService) commentService() *CommentService {
	return &CommentService{
		commentRepo: ts.commentRepo,
		postRepo:    ts.postRepo,
		videoRepo:   ts.videoRepo,
		userRepo:    ts.userRepo,
		actionRepo:  nil,
		userService: ts.userService,
	}
}
