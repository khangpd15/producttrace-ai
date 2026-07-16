package request

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required"`
	FullName string `json:"full_name" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=ADMIN CUSTOMER STAFF DEALER"`
}

type UpdateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required"`
	FullName string `json:"full_name" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=ADMIN CUSTOMER STAFF DEALER"`
	Status   string `json:"status" binding:"required,oneof=ACTIVE BANNED SUSPENDED PENDING"`
	Password string `json:"password,omitempty"`
}

type UpdateProfileRequest struct {
	FullName *string `json:"full_name"`
	Phone    *string `json:"phone"`
	Avatar   *string `json:"avatar"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

type SearchUserRequest struct {
	Keyword string `form:"keyword" binding:"max=255"`
	Role    string `form:"role" binding:"omitempty,oneof=ADMIN CUSTOMER STAFF DEALER"`
	Status  string `form:"status" binding:"omitempty,oneof=ACTIVE BANNED SUSPENDED PENDING"`
	Page    int    `form:"page"`
	Limit   int    `form:"limit"`
}
