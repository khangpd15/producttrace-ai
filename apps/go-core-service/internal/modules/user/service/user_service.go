package service

import (
	"context"
	"fmt"
	"regexp"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/publisher"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/types"
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
	SearchUsers(ctx context.Context, req *request.SearchUserRequest) (*response.UserListResponse, error)
	GetProfile(ctx context.Context, id string) (*response.UserResponse, error)
	UpdateProfile(ctx context.Context, actorID string, targetUserID string, req *request.UpdateProfileRequest) (*response.UserResponse, error)
}

type UserService struct {
	userRepo repository.UserRepositoryInterface
	pub      *publisher.Publisher
}

func NewUserService(userRepo repository.UserRepositoryInterface, pub *publisher.Publisher) UserServiceInterface {
	return &UserService{
		userRepo: userRepo,
		pub:      pub,
	}
}

func mapToUserResponse(user *entity.User) *response.UserResponse {
	avatar := ""
	if user.AvatarUrl != nil {
		avatar = *user.AvatarUrl
	}
	return &response.UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Phone:     user.Phone,
		FullName:  user.FullName,
		Role:      string(user.Role),
		Status:    string(user.Status),
		Avatar:    avatar,
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

func (s *UserService) SearchUsers(ctx context.Context, req *request.SearchUserRequest) (*response.UserListResponse, error) {
	// Validate keyword length
	if utf8.RuneCountInString(req.Keyword) > 255 {
		return nil, apperror.NewValidation("Keyword must be at most 255 characters")
	}

	// Normalize pagination defaults
	page := req.Page
	if page <= 0 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	users, total, err := s.userRepo.ListUsers(ctx, page, limit, req.Role, req.Status, req.Keyword)
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

func (s *UserService) GetProfile(ctx context.Context, id string) (*response.UserResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, apperror.WrapDBError(err, "User")
	}
	if user == nil {
		return nil, apperror.NewNotFound("User profile")
	}

	if user.Status != entity.StatusActive {
		return nil, apperror.NewForbidden("Account is not active")
	}

	return mapToUserResponse(user), nil
}

func (s *UserService) UpdateProfile(ctx context.Context, actorID string, targetUserID string, req *request.UpdateProfileRequest) (*response.UserResponse, error) {
	actor, err := s.userRepo.GetUserByID(ctx, actorID)
	if err != nil {
		return nil, apperror.WrapDBError(err, "Actor")
	}
	if actor == nil {
		return nil, apperror.NewNotFound("Actor")
	}

	if actor.Status != entity.StatusActive {
		return nil, apperror.NewForbidden("Account is not active")
	}

	if actorID != targetUserID && actor.Role != entity.RoleAdmin {
		return nil, apperror.NewForbidden("Unauthorized update")
	}

	user, err := s.userRepo.GetUserByID(ctx, targetUserID)
	if err != nil {
		return nil, apperror.WrapDBError(err, "User")
	}
	if user == nil {
		return nil, apperror.NewNotFound("User")
	}

	if req.FullName != nil {
		if *req.FullName == "" {
			return nil, apperror.NewValidation("Full name cannot be empty")
		}
		if utf8.RuneCountInString(*req.FullName) > 255 {
			return nil, apperror.NewValidation("Full name must be at most 255 characters")
		}
		user.FullName = *req.FullName
	}

	if req.Phone != nil {
		if *req.Phone == "" {
			return nil, apperror.NewValidation("Phone cannot be empty")
		}
		phoneRegex := `^(0|\+84)[0-9]{9}$`
		matched, _ := regexp.MatchString(phoneRegex, *req.Phone)
		if !matched {
			return nil, apperror.NewValidation("Invalid phone number")
		}

		if *req.Phone != user.Phone {
			phoneExists, err := s.userRepo.CheckPhoneExists(ctx, *req.Phone, targetUserID)
			if err != nil {
				return nil, apperror.NewInternal("database error checking phone uniqueness")
			}
			if phoneExists {
				return nil, apperror.NewConflict("Số điện thoại đã được sử dụng")
			}
		}
		user.Phone = *req.Phone
	}

	if req.Avatar != nil {
		user.AvatarUrl = req.Avatar
	}

	user.UpdatedAt = time.Now()

	updatedUser, err := s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return nil, apperror.WrapDBError(err, "User")
	}

	avatar := ""
	if updatedUser.AvatarUrl != nil {
		avatar = *updatedUser.AvatarUrl
	}

	auditContent := fmt.Sprintf("User %s updated profile of user %s. Name: %s, Phone: %s, Avatar: %s", actorID, targetUserID, updatedUser.FullName, updatedUser.Phone, avatar)
	err = s.userRepo.WriteAuditLog(ctx, auditContent, "PROFILE_UPDATE")
	if err != nil {
		fmt.Printf("failed to write audit log: %v\n", err)
	}

	if s.pub != nil {
		event := types.Event{
			EventID:       uuid.NewString(),
			EventType:     "user.profile_updated",
			EventVersion:  "1.0",
			Timestamp:     time.Now().UTC(),
			Producer:      "go-core-service",
			CorrelationID: uuid.NewString(),
			Payload: map[string]interface{}{
				"userId":    targetUserID,
				"fullName":  updatedUser.FullName,
				"phone":     updatedUser.Phone,
				"avatarUrl": avatar,
			},
		}
		err = s.pub.Publish(event)
		if err != nil {
			fmt.Printf("failed to publish profile updated event: %v\n", err)
		}
	}

	return mapToUserResponse(updatedUser), nil
}