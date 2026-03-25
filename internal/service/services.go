package service

type Services struct {
	UserService
	PostService
	AIService
	JWTService
}

var Service = new(Services)
