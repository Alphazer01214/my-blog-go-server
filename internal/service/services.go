package service

type Services struct {
	UserService
	PostService
	AIService
}

var Service = new(Services)
