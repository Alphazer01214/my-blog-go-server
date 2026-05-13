package service

type Services struct {
	UserService
	PostService
	CommentService
	AIService
}

var Service = new(Services)
