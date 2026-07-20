package service

import (
	"context"
	"errors"
	"time"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	"blog.alphazer01214.top/internal/utils"
	"blog.alphazer01214.top/pkg/cache"

	"gorm.io/gorm"
)

type UserService struct{}

// userCache 用户信息缓存（10 分钟 TTL）
var userCache *cache.RedisCache[response.UserInfo]

// getUserCache 懒初始化用户缓存
func getUserCache() *cache.RedisCache[response.UserInfo] {
	if userCache == nil {
		if rdb := global.GetRedis(); rdb != nil {
			userCache = cache.NewRedisCache[response.UserInfo](rdb, "cache:userinfo:", 10*time.Minute)
		}
	}
	return userCache
}

func (us *UserService) Register(user *entity.User, env *entity.EnvInfo) (*response.Register, error) {
	if us.isUsernameExist(user.Username) {
		return nil, errors.New("username already exist")
	}
	user.Password = utils.EncryptPassword(user.Password)

	err := global.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		profile := &entity.UserProfile{
			UserId:   user.ID,
			Username: user.Username,
		}
		if err := tx.Create(profile).Error; err != nil {
			return err
		}
		setting := &entity.UserSetting{
			UserId: user.ID,
		}
		if err := tx.Create(setting).Error; err != nil {
			return err
		}
		return nil
	})
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

	userInfo := us.toUserInfo(dbu, 0)

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
	accessToken, err := utils.GenerateAccessTokenFromClaims(accessClaims)
	if err != nil {
		return nil, err
	}
	refreshToken, err := utils.GenerateRefreshTokenFromClaims(refreshClaims)
	if err != nil {
		return nil, err
	}

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

func (us *UserService) GetUserInfoById(id uint, viewerId uint) (response.UserInfo, error) {
	// 尝试从缓存获取（仅当不需要检查关注状态时）
	if viewerId == 0 || viewerId == id {
		if uc := getUserCache(); uc != nil {
			cached, err := uc.Get(context.Background(), cache.CacheKey(id))
			if err == nil && cached != nil {
				return *cached, nil
			}
		}
	}

	usr, err := us.getUserInstanceById(id)
	if err != nil {
		return response.UserInfo{}, err
	}
	info := us.toUserInfo(usr, viewerId)

	// 写入缓存（仅当不需要检查关注状态时）
	if viewerId == 0 || viewerId == id {
		if uc := getUserCache(); uc != nil {
			_ = uc.Set(context.Background(), cache.CacheKey(id), &info)
		}
	}

	return info, nil
}

func (us *UserService) GetAllUserInfo(viewerId uint) ([]response.UserInfo, error) {
	var infos []response.UserInfo
	users, err := us.getAllUserInstance()
	if err != nil {
		return nil, err
	}
	for _, usr := range users {
		infos = append(infos, us.toUserInfo(&usr, viewerId))
	}
	return infos, nil
}

func (us *UserService) UpdateUserProfile(userId uint, req request.UserUpdateRequest) (*response.UserUpdate, error) {
	err := global.GetDB().Transaction(func(tx *gorm.DB) error {
		if req.NewUsername != "" {
			if err := tx.Model(&entity.User{}).Where("id = ?", userId).Update("username", req.NewUsername).Error; err != nil {
				return err
			}
		}
		updates := map[string]interface{}{}
		if req.NewEmail != "" {
			updates["email"] = req.NewEmail
		}
		if req.NewPhone != "" {
			updates["phone"] = req.NewPhone
		}
		if req.NewBio != "" {
			updates["bio"] = req.NewBio
		}
		if req.NewAvatar != "" {
			updates["avatar"] = req.NewAvatar
		}
		if req.NewUsername != "" {
			updates["username"] = req.NewUsername
		}
		if len(updates) > 0 {
			if err := tx.Model(&entity.UserProfile{}).Where("user_id = ?", userId).Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	dbUser, err := us.getUserInstanceById(userId)
	if err != nil {
		return nil, err
	}

	return &response.UserUpdate{
		Env:      req.Env,
		UserInfo: us.toUserInfo(dbUser, 0),
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
	if err := global.GetDB().Model(&entity.User{}).Where("id = ?", id).Update("password", hashedPassword).Error; err != nil {
		return nil, err
	}
	return &response.UserUpdate{
		Env: env,
	}, nil
}

func (us *UserService) Follow(followerId, followingId uint) (*response.FollowStatus, error) {
	if followerId == followingId {
		return nil, errors.New("cannot follow yourself")
	}

	var targetUser entity.User
	if err := global.GetDB().Where("id = ?", followingId).First(&targetUser).Error; err != nil {
		return nil, errors.New("user not found")
	}

	var follow entity.UserFollow
	result := global.GetDB().Where("follower_id = ? AND following_id = ?", followerId, followingId).First(&follow)

	if result.RowsAffected > 0 {
		// 已关注，取消关注
		err := global.GetDB().Transaction(func(tx *gorm.DB) error {
			if err := tx.Delete(&follow).Error; err != nil {
				return err
			}
			if follow.IsMutual {
				if err := tx.Model(&entity.UserFollow{}).
					Where("follower_id = ? AND following_id = ?", followingId, followerId).
					Update("is_mutual", false).Error; err != nil {
					return err
				}
			}
			if err := tx.Model(&entity.UserProfile{}).Where("user_id = ?", followerId).
				Update("following_count", gorm.Expr("GREATEST(following_count - 1, 0)")).Error; err != nil {
				return err
			}
			if err := tx.Model(&entity.UserProfile{}).Where("user_id = ?", followingId).
				Update("follower_count", gorm.Expr("GREATEST(follower_count - 1, 0)")).Error; err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		return &response.FollowStatus{
			IsFollowing: false,
			IsMutual:    false,
			TargetUser:  us.toUserInfo(&targetUser, followerId),
		}, nil
	}

	// 未关注，创建关注
	isMutual := false
	err := global.GetDB().Transaction(func(tx *gorm.DB) error {
		newFollow := entity.UserFollow{
			FollowerId:  followerId,
			FollowingId: followingId,
		}

		// 检查是否为互关
		var reverse entity.UserFollow
		if err := tx.Where("follower_id = ? AND following_id = ?", followingId, followerId).First(&reverse).Error; err == nil {
			newFollow.IsMutual = true
			isMutual = true
			if err := tx.Model(&reverse).Update("is_mutual", true).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(&newFollow).Error; err != nil {
			return err
		}
		if err := tx.Model(&entity.UserProfile{}).Where("user_id = ?", followerId).
			Update("following_count", gorm.Expr("following_count + 1")).Error; err != nil {
			return err
		}
		if err := tx.Model(&entity.UserProfile{}).Where("user_id = ?", followingId).
			Update("follower_count", gorm.Expr("follower_count + 1")).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// 发布 Kafka 事件：关注
	go publishFollowNotification(followerId, followingId)

	return &response.FollowStatus{
		IsFollowing: true,
		IsMutual:    isMutual,
		TargetUser:  us.toUserInfo(&targetUser, followerId),
	}, nil
}

func (us *UserService) GetFollowers(userId uint, viewerId uint, page, pageSize int) (response.FollowList, error) {
	var total int64
	if err := global.GetDB().Model(&entity.UserFollow{}).Where("following_id = ?", userId).Count(&total).Error; err != nil {
		return response.FollowList{}, err
	}

	var follows []entity.UserFollow
	offset := (page - 1) * pageSize
	if err := global.GetDB().Where("following_id = ?", userId).Order("created_at desc").Offset(offset).Limit(pageSize).Find(&follows).Error; err != nil {
		return response.FollowList{}, err
	}

	items := make([]response.UserInfo, len(follows))
	for i, f := range follows {
		usr, err := us.getUserInstanceById(f.FollowerId)
		if err != nil {
			continue
		}
		items[i] = us.toUserInfo(usr, viewerId)
	}

	return response.FollowList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (us *UserService) GetFollowing(userId uint, viewerId uint, page, pageSize int) (response.FollowList, error) {
	var total int64
	if err := global.GetDB().Model(&entity.UserFollow{}).Where("follower_id = ?", userId).Count(&total).Error; err != nil {
		return response.FollowList{}, err
	}

	var follows []entity.UserFollow
	offset := (page - 1) * pageSize
	if err := global.GetDB().Where("follower_id = ?", userId).Order("created_at desc").Offset(offset).Limit(pageSize).Find(&follows).Error; err != nil {
		return response.FollowList{}, err
	}

	items := make([]response.UserInfo, len(follows))
	for i, f := range follows {
		usr, err := us.getUserInstanceById(f.FollowingId)
		if err != nil {
			continue
		}
		items[i] = us.toUserInfo(usr, viewerId)
	}

	return response.FollowList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (us *UserService) GetSetting(userId uint) (*entity.UserSetting, error) {
	var setting entity.UserSetting
	if err := global.GetDB().Where("user_id = ?", userId).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

func (us *UserService) UpdateSetting(userId uint, req request.UserSettingUpdateRequest) error {
	if req.PostPublic == nil {
		return nil
	}
	return global.GetDB().Model(&entity.UserSetting{}).Where("user_id = ?", userId).
		Update("post_public", *req.PostPublic).Error
}

func (us *UserService) DeleteUserById(id int) error {
	return global.DB.Delete(&entity.User{}, id).Error
}

// toUserInfo 转换为 response.UserInfo，数据脱敏
func (us *UserService) toUserInfo(usr *entity.User, viewerId uint) response.UserInfo {
	var profile entity.UserProfile
	if err := global.GetDB().Where("user_id = ?", usr.ID).First(&profile).Error; err != nil {
		profile = entity.UserProfile{}
	}

	var isFollowed bool
	if viewerId > 0 && viewerId != usr.ID {
		var follow entity.UserFollow
		if err := global.GetDB().Where("follower_id = ? AND following_id = ?", viewerId, usr.ID).First(&follow).Error; err == nil {
			isFollowed = true
		}
	}

	return response.UserInfo{
		UserId:               usr.ID,
		CreatedAt:            usr.CreatedAt,
		UpdatedAt:            usr.UpdatedAt,
		Username:             usr.Username,
		Email:                profile.Email,
		Phone:                profile.Phone,
		Bio:                  profile.Bio,
		Avatar:               profile.Avatar,
		Admin:                usr.Admin,
		Role:                 usr.Role,
		Banned:               usr.Banned,
		FollowerCount:        profile.FollowerCount,
		FollowingCount:       profile.FollowingCount,
		PostCount:            profile.PostCount,
		CommentCount:         profile.CommentCount,
		ReceivedLikeCount:    profile.ReceivedLikeCount,
		ReceivedDislikeCount: profile.ReceivedDislikeCount,
		IsFollowed:           isFollowed,
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
