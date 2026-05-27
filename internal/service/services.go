package service

type Services struct {
	UserService
	PostService
	CommentService
	AIService
	MarketService
	FileService
	TomoriService
	VideoService
}

var Service = new(Services)
