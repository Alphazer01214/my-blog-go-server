package service

import (
	"errors"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	"blog.alphazer01214.top/internal/utils"
)

type UserService struct{}

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
	tokenResponse, err := us.GenerateToken(dbu)
	if err != nil {
		return nil, err
	}

	userInfo := us.toUserInfo(dbu)

	return &response.Login{
		UserInfo: userInfo,
		Token:    tokenResponse,
		Env:      env,
	}, nil
}

func (us *UserService) GenerateToken(user *entity.User) (*response.Token, error) {
	baseClaims := request.BaseClaims{
		UserId:   user.ID,
		Username: user.Username,
		RoleType: user.Role,
	}

	accessClaims := utils.GenerateAccessClaims(baseClaims)
	refreshClaims := utils.GenerateRefreshClaims(baseClaims)
	accessToken := utils.GenerateAccessTokenFromClaims(accessClaims)
	refreshToken := utils.GenerateRefreshTokenFromClaims(refreshClaims)

	refreshTokenRecord, _ := utils.GetRefreshTokenRedis(user.ID)

	if refreshTokenRecord != "" {
		if err := utils.TokenJoinBlacklist(refreshTokenRecord); err != nil {
			return nil, err
		}
	}

	if err := utils.SetRefreshTokenRedis(user.ID, refreshToken); err != nil {
		return nil, err
	}

	rp := &response.Token{
		AccessToken:            accessToken,
		RefreshToken:           refreshToken,
		AccessTokenExpireTime:  global.GetConfig().JWT.AccessTokenExpireTime,
		RefreshTokenExpireTime: global.GetConfig().JWT.RefreshTokenExpireTime,
	}

	return rp, nil
}

func (us *UserService) GetUserInfoById(id uint) (response.UserInfo, error) {
	usr, err := us.getUserInstanceById(id)
	if err != nil {
		return response.UserInfo{}, err
	}
	return us.toUserInfo(usr), nil
}

func (us *UserService) GetAllUserInfo() ([]response.UserInfo, error) {
	var infos []response.UserInfo
	users, err := us.getAllUserInstance()
	if err != nil {
		return nil, err
	}
	for _, usr := range users {
		infos = append(infos, us.toUserInfo(&usr))
	}
	return infos, nil
}

func (us *UserService) UpdateUserProfile(userId uint, req request.UserUpdateRequest) (*response.UserUpdate, error) {
	user := &entity.User{
		Username: req.NewUsername,
		Email:    req.NewEmail,
		Phone:    req.NewPhone,
		Bio:      req.NewBio,
		Avatar:   req.NewAvatar,
	}
	user.ID = userId
	if err := us.update(user); err != nil {
		return nil, err
	}
	dbUser, err := us.getUserInstanceById(userId)
	if err != nil {
		return nil, err
	}

	return &response.UserUpdate{
		Env:      req.Env,
		UserInfo: us.toUserInfo(dbUser),
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

func (us *UserService) GetPostsByUser(userId uint, page, pageSize int) (response.PostList, error) {
	return Service.PostService.GetPostsByUser(userId, page, pageSize)
}

func (us *UserService) GetCommentsByUser(userId uint, viewerId uint, page, pageSize int) ([]response.Comment, int64, error) {
	var comments []entity.Comment
	var total int64
	db := global.GetDB().Model(&entity.Comment{}).Where("user_id = ?", userId)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	author, err := Service.UserService.GetUserInfoById(userId)
	if err != nil {
		return nil, 0, err
	}

	likedIds := make(map[uint]bool)
	dislikedIds := make(map[uint]bool)
	if viewerId > 0 && len(comments) > 0 {
		allIds := make([]uint, len(comments))
		for i, c := range comments {
			allIds[i] = c.ID
		}
		var likes []entity.CommentLike
		global.GetDB().Where("comment_id IN ? AND user_id = ?", allIds, viewerId).Find(&likes)
		for _, l := range likes {
			likedIds[l.CommentId] = true
		}
		var dislikes []entity.CommentDislike
		global.GetDB().Where("comment_id IN ? AND user_id = ?", allIds, viewerId).Find(&dislikes)
		for _, d := range dislikes {
			dislikedIds[d.CommentId] = true
		}
	}

	result := make([]response.Comment, len(comments))
	for i, c := range comments {
		result[i] = response.Comment{
			CommentId:       c.ID,
			UserId:          c.UserId,
			PostId:          c.PostId,
			RootCommentId:   c.RootCommentId,
			ParentCommentId: c.ParentCommentId,
			Content:         c.Content,
			CreatedAt:       c.CreatedAt,
			UpdatedAt:       c.UpdatedAt,
			Author:          author,
			Likes:           c.Likes,
			Dislikes:        c.Dislikes,
			Replies:         c.Replies,
			IsLiked:         likedIds[c.ID],
			IsDisliked:      dislikedIds[c.ID],
		}
	}
	return result, total, nil
}

func (us *UserService) DeleteUserById(id int) error {
	return global.DB.Delete(&entity.User{}, id).Error
}

// toUserInfo 转换为 response.UserInfo，数据脱敏
func (us *UserService) toUserInfo(usr *entity.User) response.UserInfo {
	var postCount int64
	global.GetDB().Model(&entity.Post{}).Where("user_id = ?", usr.ID).Count(&postCount)
	var commentCount int64
	global.GetDB().Model(&entity.Comment{}).Where("user_id = ?", usr.ID).Count(&commentCount)

	return response.UserInfo{
		UserId:       usr.ID,
		CreatedAt:    usr.CreatedAt,
		UpdatedAt:    usr.UpdatedAt,
		Username:     usr.Username,
		Email:        usr.Email,
		Phone:        usr.Phone,
		Bio:          usr.Bio,
		Avatar:       usr.Avatar,
		Admin:        usr.Admin,
		Role:         usr.Role,
		PostCount:    postCount,
		CommentCount: commentCount,
	}
}

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
	if err := global.DB.Where("id = ?", id).First(&user).Error; err != nil {
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

func (us *UserService) update(user *entity.User) error {
	return global.DB.Model(&entity.User{}).Where("id = ?", user.ID).Updates(user).Error
}
