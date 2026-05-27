package service

import (
	"blog.alphazer01214.top/internal/repository"
)

type Services struct {
	UserService    *UserService
	PostService    *PostService
	CommentService *CommentService
	AIService      *AIService
	MarketService  *MarketService
	FileService    *FileService
	TomoriService  *TomoriService
	VideoService   *VideoService
}

func NewServices(repos *repository.Repositories) *Services {
	userSvc := NewUserService(repos.User)
	postSvc := NewPostService(repos.Post, repos.User, repos.Action, userSvc)
	commentSvc := NewCommentService(repos.Comment, repos.Post, repos.Video, repos.User, repos.Action, userSvc)
	videoSvc := NewVideoService(repos.Video, repos.Action, repos.User, userSvc)
	aiSvc := NewAIService(repos.AI)
	fileSvc := NewFileService(repos.Upload, repos.Video, userSvc)
	tomoriSvc := NewTomoriService(repos.User, repos.Post, repos.Comment, repos.Video, repos.AI, userSvc, postSvc, fileSvc)

	return &Services{
		UserService:    userSvc,
		PostService:    postSvc,
		CommentService: commentSvc,
		AIService:      aiSvc,
		MarketService:  &MarketService{},
		FileService:    fileSvc,
		TomoriService:  tomoriSvc,
		VideoService:   videoSvc,
	}
}
