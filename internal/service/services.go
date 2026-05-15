package service

type Services struct {
	UserService
	PostService
	CommentService
	AIService
	MarketService
}

var Service = new(Services)
