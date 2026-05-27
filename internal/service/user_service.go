package service

import (
	"context"
	"errors"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/repository"
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	"blog.alphazer01214.top/internal/utils"

	"gorm.io/gorm"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// ======================== Auth ========================

func (us *UserService) Register(ctx context.Context, user *entity.User, env *entity.EnvInfo) (*response.Register, error) {
	exists, err := us.repo.ExistsByUsername(ctx, user.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("username already exist")
	}
	user.Password = utils.EncryptPassword(user.Password)

	err = us.repo.WithTransaction(ctx, func(tx *gorm.DB) error {
		txRepo := us.repo.WithTx(tx)
		if err := txRepo.Create(ctx, user); err != nil {
			return err
		}
		if err := txRepo.CreateProfile(ctx, &entity.UserProfile{
			UserId:   user.ID,
			Username: user.Username,
		}); err != nil {
			return err
		}
		return txRepo.CreateSetting(ctx, &entity.UserSetting{
			UserId: user.ID,
		})
	})
	if err != nil {
		return nil, err
	}
	return &response.Register{
		Username: user.Username,
		Env:      env,
	}, nil
}

func (us *UserService) Login(ctx context.Context, user *entity.User, env *entity.EnvInfo) (*response.Login, error) {
	dbUser, err := us.repo.FindByUsername(ctx, user.Username)
	if err != nil {
		return nil, errors.New("user not exist")
	}
	if !utils.IsPasswordCorrect(user.Password, dbUser.Password) {
		return nil, errors.New("wrong password")
	}
	if dbUser.Banned {
		return nil, errors.New("user banned")
	}
	tokenResponse, err := us.GenerateToken(ctx, dbUser)
	if err != nil {
		return nil, err
	}
	userInfo, err := us.toUserInfo(ctx, dbUser, 0)
	if err != nil {
		return nil, err
	}
	return &response.Login{
		UserInfo: userInfo,
		Token:    tokenResponse,
		Env:      env,
	}, nil
}

func (us *UserService) GenerateToken(ctx context.Context, user *entity.User) (*response.Token, error) {
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

	return &response.Token{
		AccessToken:            accessToken,
		RefreshToken:           refreshToken,
		AccessTokenExpireTime:  global.GetConfig().JWT.AccessTokenExpireTime,
		RefreshTokenExpireTime: global.GetConfig().JWT.RefreshTokenExpireTime,
	}, nil
}

// ======================== User Query ========================

func (us *UserService) GetUserById(ctx context.Context, id uint, viewerId uint) (*response.UserInfo, error) {
	user, err := us.repo.FindById(ctx, id)
	if err != nil {
		return nil, err
	}
	info, err := us.toUserInfo(ctx, user, viewerId)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (us *UserService) ListUsers(ctx context.Context, viewerId uint) (*response.UserList, error) {
	users, err := us.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]response.UserInfo, len(users))
	for i, u := range users {
		info, _ := us.toUserInfo(ctx, &u, viewerId)
		items[i] = info
	}
	return &response.UserList{Items: items}, nil
}

// ======================== Profile Update ========================

func (us *UserService) UpdateProfile(ctx context.Context, userId uint, req request.UserUpdateRequest) (*response.UserUpdate, error) {
	err := us.repo.WithTransaction(ctx, func(tx *gorm.DB) error {
		txRepo := us.repo.WithTx(tx)
		if req.NewUsername != "" {
			if err := txRepo.UpdateUsername(ctx, userId, req.NewUsername); err != nil {
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
			return txRepo.UpdateProfileFields(ctx, userId, updates)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	dbUser, err := us.repo.FindById(ctx, userId)
	if err != nil {
		return nil, err
	}
	info, err := us.toUserInfo(ctx, dbUser, 0)
	if err != nil {
		return nil, err
	}
	return &response.UserUpdate{
		Env:      req.Env,
		UserInfo: info,
	}, nil
}

func (us *UserService) UpdatePassword(ctx context.Context, id uint, oldPassword, newPassword string, env *entity.EnvInfo) (*response.UserUpdate, error) {
	if !utils.IsPasswordValid(oldPassword) || !utils.IsPasswordValid(newPassword) {
		return nil, errors.New("password invalid")
	}
	dbPassword, err := us.repo.GetPassword(ctx, id)
	if err != nil {
		return nil, errors.New("wrong password")
	}
	if !utils.IsPasswordCorrect(oldPassword, dbPassword) {
		return nil, errors.New("wrong password")
	}
	hashedPassword := utils.EncryptPassword(newPassword)
	if err := us.repo.UpdatePassword(ctx, id, hashedPassword); err != nil {
		return nil, err
	}
	return &response.UserUpdate{Env: env}, nil
}

// ======================== Follow ========================

func (us *UserService) Follow(ctx context.Context, followerId, followingId uint) (*response.FollowStatus, error) {
	if followerId == followingId {
		return nil, errors.New("cannot follow yourself")
	}
	targetUser, err := us.repo.FindById(ctx, followingId)
	if err != nil {
		return nil, errors.New("user not found")
	}

	follow, err := us.repo.FindFollow(ctx, followerId, followingId)
	if err == nil && follow != nil {
		// Already following → unfollow
		err = us.repo.WithTransaction(ctx, func(tx *gorm.DB) error {
			txRepo := us.repo.WithTx(tx)
			if err := txRepo.DeleteFollow(ctx, followerId, followingId); err != nil {
				return err
			}
			if follow.IsMutual {
				if err := txRepo.UpdateFollowMutual(ctx, followingId, followerId, false); err != nil {
					return err
				}
			}
			if err := txRepo.DecrementProfileCounter(ctx, followerId, "following_count"); err != nil {
				return err
			}
			return txRepo.DecrementProfileCounter(ctx, followingId, "follower_count")
		})
		if err != nil {
			return nil, err
		}
		info, _ := us.toUserInfo(ctx, targetUser, followerId)
		return &response.FollowStatus{
			IsFollowing: false,
			IsMutual:    false,
			TargetUser:  info,
		}, nil
	}

	// Not following → create follow
	isMutual := false
	err = us.repo.WithTransaction(ctx, func(tx *gorm.DB) error {
		txRepo := us.repo.WithTx(tx)
		newFollow := entity.UserFollow{
			FollowerId:  followerId,
			FollowingId: followingId,
		}
		reverse, err := txRepo.FindFollow(ctx, followingId, followerId)
		if err == nil && reverse != nil {
			newFollow.IsMutual = true
			isMutual = true
			if err := txRepo.UpdateFollowMutual(ctx, followingId, followerId, true); err != nil {
				return err
			}
		}
		if err := txRepo.CreateFollow(ctx, &newFollow); err != nil {
			return err
		}
		if err := txRepo.IncrementProfileCounter(ctx, followerId, "following_count"); err != nil {
			return err
		}
		return txRepo.IncrementProfileCounter(ctx, followingId, "follower_count")
	})
	if err != nil {
		return nil, err
	}
	info, _ := us.toUserInfo(ctx, targetUser, followerId)
	return &response.FollowStatus{
		IsFollowing: true,
		IsMutual:    isMutual,
		TargetUser:  info,
	}, nil
}

func (us *UserService) ListFollowers(ctx context.Context, userId, viewerId uint, page, pageSize int) (*response.FollowList, error) {
	pagination, err := us.repo.ListFollowers(ctx, userId, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]response.UserInfo, len(pagination.Items))
	for i, f := range pagination.Items {
		user, err := us.repo.FindById(ctx, f.FollowerId)
		if err != nil {
			continue
		}
		info, _ := us.toUserInfo(ctx, user, viewerId)
		items[i] = info
	}
	return &response.FollowList{
		Items:    items,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Total:    pagination.Total,
	}, nil
}

func (us *UserService) ListFollowing(ctx context.Context, userId, viewerId uint, page, pageSize int) (*response.FollowList, error) {
	pagination, err := us.repo.ListFollowing(ctx, userId, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]response.UserInfo, len(pagination.Items))
	for i, f := range pagination.Items {
		user, err := us.repo.FindById(ctx, f.FollowingId)
		if err != nil {
			continue
		}
		info, _ := us.toUserInfo(ctx, user, viewerId)
		items[i] = info
	}
	return &response.FollowList{
		Items:    items,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Total:    pagination.Total,
	}, nil
}

// ======================== Settings ========================

func (us *UserService) GetSetting(ctx context.Context, userId uint) (*entity.UserSetting, error) {
	return us.repo.FindSettingByUserId(ctx, userId)
}

func (us *UserService) UpdateSetting(ctx context.Context, userId uint, req request.UserSettingUpdateRequest) error {
	if req.PostPublic == nil {
		return nil
	}
	return us.repo.UpdateSettingPostPublic(ctx, userId, *req.PostPublic)
}

func (us *UserService) DeleteUserById(ctx context.Context, id uint) error {
	return us.repo.DeleteById(ctx, id)
}

// ======================== Internal Helpers ========================

func (us *UserService) toUserInfo(ctx context.Context, user *entity.User, viewerId uint) (response.UserInfo, error) {
	profile, err := us.repo.FindProfileByUserId(ctx, user.ID)
	if err != nil {
		profile = &entity.UserProfile{}
	}

	isFollowed := false
	if viewerId > 0 && viewerId != user.ID {
		follow, err := us.repo.FindFollow(ctx, viewerId, user.ID)
		if err == nil && follow != nil {
			isFollowed = true
		}
	}

	return response.UserInfo{
		UserId:               user.ID,
		CreatedAt:            user.CreatedAt,
		UpdatedAt:            user.UpdatedAt,
		Username:             user.Username,
		Email:                profile.Email,
		Phone:                profile.Phone,
		Bio:                  profile.Bio,
		Avatar:               profile.Avatar,
		Admin:                user.Admin,
		Role:                 user.Role,
		Banned:               user.Banned,
		FollowerCount:        profile.FollowerCount,
		FollowingCount:       profile.FollowingCount,
		PostCount:            profile.PostCount,
		CommentCount:         profile.CommentCount,
		ReceivedLikeCount:    profile.ReceivedLikeCount,
		ReceivedDislikeCount: profile.ReceivedDislikeCount,
		IsFollowed:           isFollowed,
	}, nil
}
