package repository

import (
	"context"

	"blog.alphazer01214.top/internal/entity"
	"gorm.io/gorm"
)

type UserRepository interface {
	WithTx(tx *gorm.DB) UserRepository
	WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error

	// User CRUD
	Create(ctx context.Context, user *entity.User) error
	FindById(ctx context.Context, id uint) (*entity.User, error)
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	GetPassword(ctx context.Context, id uint) (string, error)
	UpdatePassword(ctx context.Context, id uint, hashedPassword string) error
	UpdateUsername(ctx context.Context, id uint, username string) error
	UpdateFields(ctx context.Context, id uint, updates map[string]interface{}) error
	UpdateBanned(ctx context.Context, id uint, banned bool) (int64, error)
	UpdateRole(ctx context.Context, id uint, role entity.RoleType) (int64, error)
	DeleteById(ctx context.Context, id uint) error
	FindAll(ctx context.Context) ([]entity.User, error)
	ListUsers(ctx context.Context, page, pageSize int) (Pagination[entity.User], error)

	// Profile
	CreateProfile(ctx context.Context, profile *entity.UserProfile) error
	FindProfileByUserId(ctx context.Context, userId uint) (*entity.UserProfile, error)
	UpdateProfileFields(ctx context.Context, userId uint, updates map[string]interface{}) error
	IncrementProfileCounter(ctx context.Context, userId uint, field string) error
	DecrementProfileCounter(ctx context.Context, userId uint, field string) error
	DeleteProfileByUserId(ctx context.Context, userId uint) error

	// Setting
	CreateSetting(ctx context.Context, setting *entity.UserSetting) error
	FindSettingByUserId(ctx context.Context, userId uint) (*entity.UserSetting, error)
	UpdateSettingPostPublic(ctx context.Context, userId uint, postPublic bool) error
	DeleteSettingByUserId(ctx context.Context, userId uint) error

	// Follow
	FindFollow(ctx context.Context, followerId, followingId uint) (*entity.UserFollow, error)
	CreateFollow(ctx context.Context, follow *entity.UserFollow) error
	DeleteFollow(ctx context.Context, followerId, followingId uint) error
	UpdateFollowMutual(ctx context.Context, followerId, followingId uint, mutual bool) error
	ListFollowers(ctx context.Context, userId uint, page, pageSize int) (Pagination[entity.UserFollow], error)
	ListFollowing(ctx context.Context, userId uint, page, pageSize int) (Pagination[entity.UserFollow], error)
	DeleteFollowsByUserId(ctx context.Context, userId uint) error

	// Token Blacklist
	AddToBlacklist(ctx context.Context, token string) error
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
	ListBlacklist(ctx context.Context, page, pageSize int) (Pagination[entity.TokenBlacklist], error)
	ClearBlacklist(ctx context.Context) error
}
