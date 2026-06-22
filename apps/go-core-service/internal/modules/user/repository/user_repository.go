package repository

import (
	"context"
	"errors"

	UserEntity "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/entity"
	"gorm.io/gorm"
)

type contextKey string
const TxKey contextKey = "gormTx"

type UserRepositoryInterface interface {
	CreateUser(ctx context.Context, user *UserEntity.User) (*UserEntity.User, error)
	GetUserByEmail(ctx context.Context, email string) (*UserEntity.User, error)
	GetUserByID(ctx context.Context, id string) (*UserEntity.User, error)
	UpdateUserStatus(ctx context.Context, id string, status UserEntity.Status) error
	CheckEmailExists(ctx context.Context, email string) (bool, error)
	UpdateUser(ctx context.Context, user *UserEntity.User) (*UserEntity.User, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, page, limit int, role, status, search string) ([]*UserEntity.User, int64, error)
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepositoryInterface {
	return &UserRepository{
		db: db,
	}
}

// getDB returns GORM DB from context (if transaction exists) or the default DB connection.
func (r *UserRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(TxKey).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}

func (r *UserRepository) CreateUser(ctx context.Context, user *UserEntity.User) (*UserEntity.User, error) {
	db := r.getDB(ctx)
	if err := db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*UserEntity.User, error) {
	db := r.getDB(ctx)
	var user UserEntity.User
	if err := db.Where("email = ? AND is_deleted = ?", email, false).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateUserStatus(ctx context.Context, id string, status UserEntity.Status) error {
	db := r.getDB(ctx)
	return db.Model(&UserEntity.User{}).Where("id = ? AND is_deleted = ?", id, false).Update("status", status).Error
}

func (r *UserRepository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	db := r.getDB(ctx)
	var count int64
	if err := db.Model(&UserEntity.User{}).Where("email = ? AND is_deleted = ?", email, false).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*UserEntity.User, error) {
	db := r.getDB(ctx)
	var user UserEntity.User
	if err := db.Where("id = ? AND is_deleted = ?", id, false).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *UserEntity.User) (*UserEntity.User, error) {
	db := r.getDB(ctx)
	if err := db.Save(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, id string) error {
	db := r.getDB(ctx)
	return db.Model(&UserEntity.User{}).Where("id = ?", id).Update("is_deleted", true).Error
}

func (r *UserRepository) ListUsers(ctx context.Context, page, limit int, role, status, search string) ([]*UserEntity.User, int64, error) {
	db := r.getDB(ctx).Model(&UserEntity.User{}).Where("is_deleted = ?", false)

	if role != "" {
		db = db.Where("role = ?", role)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if search != "" {
		db = db.Where("email LIKE ? OR phone LIKE ? OR full_name LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	var users []*UserEntity.User
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}