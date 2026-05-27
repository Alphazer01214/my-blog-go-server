package repository

import "gorm.io/gorm"

type Repositories struct {
	User    UserRepository
	Post    PostRepository
	Comment CommentRepository
	Video   VideoRepository
	Action  ActionRepository
	AI      AIRepository
	Upload  UploadRepository
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:    NewUserRepository(db),
		Post:    NewPostRepository(db),
		Comment: NewCommentRepository(db),
		Video:   NewVideoRepository(db),
		Action:  NewActionRepository(db),
		AI:      NewAIRepository(db),
		Upload:  NewUploadRepository(db),
	}
}
