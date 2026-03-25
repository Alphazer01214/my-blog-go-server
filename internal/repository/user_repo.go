package repository

import (
	"blog.alphazer01214.top/internal/entity"
	"gorm.io/gorm"
)

type UserRepository interface {
	Register(user *entity.User) error
	GetUserInstanceById(id uint) (*entity.User, error)
	GetUserInstanceByUsername(username string) (*entity.User, error)
	GetAllUserInstance() ([]entity.User, error)
	UpdateUserProfile(user *entity.User) error
	UpdateUserPassword(id uint, hashedPassword string) error
	DeleteUserById(id int) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (u *userRepository) Register(user *entity.User) error {
	return u.db.Create(user).Error
}

func (u *userRepository) GetUserInstanceById(id uint) (*entity.User, error) {
	var user entity.User
	err := u.db.First(&user, id).Error
	return &user, err
}

func (u *userRepository) GetUserInstanceByUsername(username string) (*entity.User, error) {
	var user entity.User
	err := u.db.Where("username = ?", username).First(&user).Error
	return &user, err
}
func (u *userRepository) GetAllUserInstance() ([]entity.User, error) {
	var users []entity.User
	err := u.db.Find(&users).Error
	return users, err
}

func (u *userRepository) UpdateUserProfile(user *entity.User) error {
	return u.db.Updates(user).Error
}

func (u *userRepository) UpdateUserPassword(id uint, hashedPassword string) error {
	return u.db.Model(&entity.User{}).Where("id = ?", id).Update("password", hashedPassword).Error
}

func (u *userRepository) DeleteUserById(id int) error {
	return u.db.Delete(&entity.User{}, id).Error
}
