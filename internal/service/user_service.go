package service

import (
	"errors"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/response"
	"blog.alphazer01214.top/internal/utils"
)

type UserService struct{}

func NewUserService() UserService {
	return UserService{}
}

func (us *UserService) Register(user *entity.User, env *entity.EnvInfo) (*response.Register, error) {
	if us.isUsernameExist(user.Username) {
		return nil, errors.New("username already exist")
	}
	user.Password = utils.EncryptPassword(user.Password)
	err := us.add(user)
	if err != nil {
		return nil, err
	}
	return &response.Register{
		Username: user.Username,
		Env:      env,
	}, nil
}

func (us *UserService) Login(user *entity.User, env *entity.EnvInfo) (*response.Login, error) {
	var dbu *entity.User
	dbu, err := us.getUserInstanceByUsername(user.Username)
	if err != nil {
		return nil, errors.New("user not exist")
	}
	if !utils.IsPasswordCorrect(user.Password, dbu.Password) {
		return nil, errors.New("wrong password")
	}
	if dbu.Banned {
		return nil, errors.New("user banned")
	}

	// 生成 JWT token
	token, err := utils.GenerateToken(dbu.ID, dbu.Username)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}
	global.User.Store(dbu.ID, token)

	rp := &response.Login{
		UserInfo:  user,
		Token:     token,
		ExpiresIn: 1145141919810,
		Env:       env,
	}
	return rp, nil
}

func (us *UserService) Logout(user *entity.User) error {
	if _, ok := global.User.Load(user.ID); ok {
		global.User.Delete(user.ID)
		return nil
	}

	return errors.New("user not login")
}

func (us *UserService) GetUserById(id uint) (*response.UserQueryOne, error) {
	user, err := us.getUserInstanceById(id)
	if err != nil {
		return nil, err
	}
	return &response.UserQueryOne{
		UserInfo: user,
	}, nil
}

func (us *UserService) GetAllUserInstance() (*response.UserQueryMultiple, error) {
	users, err := us.getAllUserInstance()
	if err != nil {
		return nil, err
	}
	return &response.UserQueryMultiple{
		Users: users,
	}, nil
}

func (us *UserService) UpdateUserProfile(user *entity.User, env *entity.EnvInfo) (*response.UserUpdate, error) {
	if err := us.update(user); err != nil {
		return nil, err
	}

	return &response.UserUpdate{
		Env:      env,
		UserInfo: user,
	}, nil
}

func (us *UserService) UpdateUserPassword(id uint, oldPassword string, newPassword string, env *entity.EnvInfo) (*response.UserUpdate, error) {
	if !utils.IsPasswordValid(oldPassword) || !utils.IsPasswordValid(newPassword) {
		return nil, errors.New("password invalid")
	}
	if !us.isPasswordCorrect(id, oldPassword) {
		return nil, errors.New("wrong password")
	}
	hashedPassword := utils.EncryptPassword(newPassword)
	if err := global.DB.Model(&entity.User{}).Where("id = ?", id).Update("password", hashedPassword).Error; err != nil {
		return nil, err
	}
	return &response.UserUpdate{
		Env: env,
	}, nil
}

func (us *UserService) DeleteUserById(id int) error {
	return global.DB.Delete(&entity.User{}, id).Error
}

// Need private method to manage db

func (us *UserService) add(user *entity.User) error {
	return global.DB.Create(user).Error
}

func (us *UserService) isUsernameExist(username string) bool {
	return global.DB.Where("username = ?", username).First(&entity.User{}).Error == nil
}

func (us *UserService) isPasswordCorrect(id uint, clear string) bool {
	var dbp string
	err := global.DB.Model(&entity.User{}).Where("id = ?", id).Pluck("password", &dbp).Error
	if err != nil {
		return false
	}
	return utils.IsPasswordCorrect(clear, dbp)
}

func (us *UserService) getUserInstanceById(id uint) (*entity.User, error) {
	var user entity.User
	if err := global.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (us *UserService) getUserInstanceByUsername(username string) (*entity.User, error) {
	var user entity.User
	if err := global.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (us *UserService) getAllUserInstance() ([]entity.User, error) {
	var users []entity.User
	if err := global.DB.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// update: update all, including password
func (us *UserService) update(user *entity.User) error {
	return global.DB.Model(&entity.User{}).Where("id = ?", user.ID).Updates(user).Error
}
