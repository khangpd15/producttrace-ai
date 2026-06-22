package service

import (
	"context"
	"time"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/entity"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/repository"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/utils"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
)

type UserServiceInterface interface {
	CreateUser(ctx context.Context, req *request.CreateUserRequest) (*response.UserResponse, error)
	GetUserByID(ctx context.Context, id string) (*response.UserResponse, error)
	UpdateUser(ctx context.Context, id string, req *request.UpdateUserRequest) (*response.UserResponse, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, page, limit int, role, status, search string) (*response.UserListResponse, error)
	UpdateProfile(ctx context.Context, id string, req *request.UpdateProfileRequest) (*response.UserResponse, error)
}

type UserService struct {
	userRepo repository.UserRepositoryInterface
}

func NewUserService(userRepo repository.UserRepositoryInterface) UserServiceInterface {
	return &UserService{
		userRepo: userRepo,
	}
}

func mapToUserResponse(user *entity.User) *response.UserResponse {
	return &response.UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Phone:     user.Phone,
		FullName:  user.FullName,
		Role:      string(user.Role),
		Status:    string(user.Status),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *request.CreateUserRequest) (*response.UserResponse, error) {
	exists, err := s.userRepo.CheckEmailExists(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.NewConflict("User with this email already exists")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	newUser := entity.NewUser(req.Email, req.Phone, req.FullName, hashedPassword, req.Role)
	savedUser, err := s.userRepo.CreateUser(ctx, newUser)
	if err != nil {
		return nil, err
	}

	return mapToUserResponse(savedUser), nil
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*response.UserResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperror.NewNotFound("User")
	}

	return mapToUserResponse(user), nil
}

func (s *UserService) UpdateUser(ctx context.Context, id string, req *request.UpdateUserRequest) (*response.UserResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperror.NewNotFound("User")
	}

	if req.Email != user.Email {
		exists, err := s.userRepo.CheckEmailExists(ctx, req.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperror.NewConflict("User with this email already exists")
		}
	}

	user.Email = req.Email
	user.Phone = req.Phone
	user.FullName = req.FullName
	user.Role = entity.Role(req.Role)
	user.Status = entity.Status(req.Status)
	user.UpdatedAt = time.Now()

	if req.Password != "" {
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = hashedPassword
	}

	updatedUser, err := s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return mapToUserResponse(updatedUser), nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return apperror.NewNotFound("User")
	}

	return s.userRepo.DeleteUser(ctx, id)
}

func (s *UserService) ListUsers(ctx context.Context, page, limit int, role, status, search string) (*response.UserListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	users, total, err := s.userRepo.ListUsers(ctx, page, limit, role, status, search)
	if err != nil {
		return nil, err
	}

	items := make([]*response.UserResponse, len(users))
	for i, u := range users {
		items[i] = mapToUserResponse(u)
	}

	return &response.UserListResponse{
		Items: items,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, id string, req *request.UpdateProfileRequest) (*response.UserResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperror.NewNotFound("User")
	}

	user.FullName = req.FullName
	user.Phone = req.Phone
	user.UpdatedAt = time.Now()

	updatedUser, err := s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return mapToUserResponse(updatedUser), nil
}