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
	SignalService
	WatchlistService
}

var Service = new(Services)
