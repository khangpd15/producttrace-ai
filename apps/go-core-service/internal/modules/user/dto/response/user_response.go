package response

import "time"

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserListResponse struct {
	Items []*UserResponse `json:"items"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
}
