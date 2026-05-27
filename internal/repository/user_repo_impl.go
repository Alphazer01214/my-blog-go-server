package repository

import (
	"context"

	"blog.alphazer01214.top/internal/entity"
	"gorm.io/gorm"
)

type UserRepositoryImpl struct {
	*BaseRepository
}

func NewUserRepository(db *gorm.DB) *UserRepositoryImpl {
	return &UserRepositoryImpl{BaseRepository: NewBaseRepository(db)}
}

func (r *UserRepositoryImpl) WithTx(tx *gorm.DB) UserRepository {
	return &UserRepositoryImpl{BaseRepository: &BaseRepository{DB: tx}}
}

// ======================== User CRUD ========================

func (r *UserRepositoryImpl) Create(ctx context.Context, user *entity.User) error {
	return r.DB.WithContext(ctx).Create(user).Error
}

func (r *UserRepositoryImpl) FindById(ctx context.Context, id uint) (*entity.User, error) {
	var user entity.User
	err := r.DB.WithContext(ctx).Where("id = ?", id).First(&user).Error
	return &user, err
}

func (r *UserRepositoryImpl) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	err := r.DB.WithContext(ctx).Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *UserRepositoryImpl) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.DB.WithContext(ctx).Model(&entity.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

func (r *UserRepositoryImpl) GetPassword(ctx context.Context, id uint) (string, error) {
	var password string
	err := r.DB.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Pluck("password", &password).Error
	return password, err
}

func (r *UserRepositoryImpl) UpdatePassword(ctx context.Context, id uint, hashedPassword string) error {
	return r.DB.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Update("password", hashedPassword).Error
}

func (r *UserRepositoryImpl) UpdateUsername(ctx context.Context, id uint, username string) error {
	return r.DB.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Update("username", username).Error
}

func (r *UserRepositoryImpl) UpdateFields(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.DB.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Updates(updates).Error
}

func (r *UserRepositoryImpl) UpdateBanned(ctx context.Context, id uint, banned bool) (int64, error) {
	result := r.DB.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Update("banned", banned)
	return result.RowsAffected, result.Error
}

func (r *UserRepositoryImpl) UpdateRole(ctx context.Context, id uint, role entity.RoleType) (int64, error) {
	result := r.DB.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Update("role", role)
	return result.RowsAffected, result.Error
}

func (r *UserRepositoryImpl) DeleteById(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Delete(&entity.User{}, id).Error
}

func (r *UserRepositoryImpl) FindAll(ctx context.Context) ([]entity.User, error) {
	var users []entity.User
	err := r.DB.WithContext(ctx).Find(&users).Error
	return users, err
}

func (r *UserRepositoryImpl) ListUsers(ctx context.Context, page, pageSize int) (Pagination[entity.User], error) {
	db := r.DB.WithContext(ctx).Model(&entity.User{})
	return Paginate[entity.User](db, page, pageSize, "id desc")
}

// ======================== Profile ========================

func (r *UserRepositoryImpl) CreateProfile(ctx context.Context, profile *entity.UserProfile) error {
	return r.DB.WithContext(ctx).Create(profile).Error
}

func (r *UserRepositoryImpl) FindProfileByUserId(ctx context.Context, userId uint) (*entity.UserProfile, error) {
	var profile entity.UserProfile
	err := r.DB.WithContext(ctx).Where("user_id = ?", userId).First(&profile).Error
	return &profile, err
}

func (r *UserRepositoryImpl) UpdateProfileFields(ctx context.Context, userId uint, updates map[string]interface{}) error {
	return r.DB.WithContext(ctx).Model(&entity.UserProfile{}).Where("user_id = ?", userId).Updates(updates).Error
}

func (r *UserRepositoryImpl) IncrementProfileCounter(ctx context.Context, userId uint, field string) error {
	return r.DB.WithContext(ctx).
		Model(&entity.UserProfile{}).
		Where("user_id = ?", userId).
		Update(field, gorm.Expr(field+" + 1")).Error
}

func (r *UserRepositoryImpl) DecrementProfileCounter(ctx context.Context, userId uint, field string) error {
	return r.DB.WithContext(ctx).
		Model(&entity.UserProfile{}).
		Where("user_id = ?", userId).
		Update(field, gorm.Expr("GREATEST("+field+" - 1, 0)")).Error
}

func (r *UserRepositoryImpl) DeleteProfileByUserId(ctx context.Context, userId uint) error {
	return r.DB.WithContext(ctx).Where("user_id = ?", userId).Delete(&entity.UserProfile{}).Error
}

// ======================== Setting ========================

func (r *UserRepositoryImpl) CreateSetting(ctx context.Context, setting *entity.UserSetting) error {
	return r.DB.WithContext(ctx).Create(setting).Error
}

func (r *UserRepositoryImpl) FindSettingByUserId(ctx context.Context, userId uint) (*entity.UserSetting, error) {
	var setting entity.UserSetting
	err := r.DB.WithContext(ctx).Where("user_id = ?", userId).First(&setting).Error
	return &setting, err
}

func (r *UserRepositoryImpl) UpdateSettingPostPublic(ctx context.Context, userId uint, postPublic bool) error {
	return r.DB.WithContext(ctx).
		Model(&entity.UserSetting{}).
		Where("user_id = ?", userId).
		Update("post_public", postPublic).Error
}

func (r *UserRepositoryImpl) DeleteSettingByUserId(ctx context.Context, userId uint) error {
	return r.DB.WithContext(ctx).Where("user_id = ?", userId).Delete(&entity.UserSetting{}).Error
}

// ======================== Follow ========================

func (r *UserRepositoryImpl) FindFollow(ctx context.Context, followerId, followingId uint) (*entity.UserFollow, error) {
	var follow entity.UserFollow
	err := r.DB.WithContext(ctx).
		Where("follower_id = ? AND following_id = ?", followerId, followingId).
		First(&follow).Error
	return &follow, err
}

func (r *UserRepositoryImpl) CreateFollow(ctx context.Context, follow *entity.UserFollow) error {
	return r.DB.WithContext(ctx).Create(follow).Error
}

func (r *UserRepositoryImpl) DeleteFollow(ctx context.Context, followerId, followingId uint) error {
	return r.DB.WithContext(ctx).
		Where("follower_id = ? AND following_id = ?", followerId, followingId).
		Delete(&entity.UserFollow{}).Error
}

func (r *UserRepositoryImpl) UpdateFollowMutual(ctx context.Context, followerId, followingId uint, mutual bool) error {
	return r.DB.WithContext(ctx).
		Model(&entity.UserFollow{}).
		Where("follower_id = ? AND following_id = ?", followerId, followingId).
		Update("is_mutual", mutual).Error
}

func (r *UserRepositoryImpl) ListFollowers(ctx context.Context, userId uint, page, pageSize int) (Pagination[entity.UserFollow], error) {
	db := r.DB.WithContext(ctx).
		Model(&entity.UserFollow{}).
		Where("following_id = ?", userId)
	return Paginate[entity.UserFollow](db, page, pageSize, "created_at desc")
}

func (r *UserRepositoryImpl) ListFollowing(ctx context.Context, userId uint, page, pageSize int) (Pagination[entity.UserFollow], error) {
	db := r.DB.WithContext(ctx).
		Model(&entity.UserFollow{}).
		Where("follower_id = ?", userId)
	return Paginate[entity.UserFollow](db, page, pageSize, "created_at desc")
}

func (r *UserRepositoryImpl) DeleteFollowsByUserId(ctx context.Context, userId uint) error {
	return r.DB.WithContext(ctx).
		Where("follower_id = ? OR following_id = ?", userId, userId).
		Delete(&entity.UserFollow{}).Error
}

// ======================== Token Blacklist ========================

func (r *UserRepositoryImpl) AddToBlacklist(ctx context.Context, token string) error {
	return r.DB.WithContext(ctx).Create(&entity.TokenBlacklist{Token: token}).Error
}

func (r *UserRepositoryImpl) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	var count int64
	err := r.DB.WithContext(ctx).
		Model(&entity.TokenBlacklist{}).
		Where("token = ?", token).
		Count(&count).Error
	return count > 0, err
}

func (r *UserRepositoryImpl) ListBlacklist(ctx context.Context, page, pageSize int) (Pagination[entity.TokenBlacklist], error) {
	db := r.DB.WithContext(ctx).Model(&entity.TokenBlacklist{})
	return Paginate[entity.TokenBlacklist](db, page, pageSize, "created_at desc")
}

func (r *UserRepositoryImpl) ClearBlacklist(ctx context.Context) error {
	return r.DB.WithContext(ctx).Where("1 = 1").Delete(&entity.TokenBlacklist{}).Error
}
