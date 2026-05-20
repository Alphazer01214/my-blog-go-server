package service

type Services struct {
	UserService
	PostService
	CommentService
	AIService
	MarketService
	FileService
	TomoriService
}

var Service = new(Services)
