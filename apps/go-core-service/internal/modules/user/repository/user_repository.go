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
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateUserStatus(ctx context.Context, id string, status UserEntity.Status) error {
	db := r.getDB(ctx)
	return db.Model(&UserEntity.User{}).Where("id = ?", id).Update("status", status).Error
}

func (r *UserRepository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	db := r.getDB(ctx)
	var count int64
	if err := db.Model(&UserEntity.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*UserEntity.User, error) {
	db := r.getDB(ctx)
	var user UserEntity.User
	if err := db.Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}