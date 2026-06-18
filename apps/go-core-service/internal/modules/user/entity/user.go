package user

import (
	"time"

	"github.com/google/uuid"
)
type Role string
type Status string
const (
	RoleAdmin Role = "admin"
	RoleCustomer Role = "customer"
	RoleStaff Role = "staff"
	RoleDealer Role = "dealer"
)

const (
	StatusActive Status = "active"
	StatusBanned Status = "banned"
	StatusSuspended Status = "suspended"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	FullName     string    `json:"full_name"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	Status       Status    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsDeleted    bool      `json:"is_deleted"`
}

func NewUser(email, phone, full_name, password_hash, role string) *User {
	return &User{
		ID: uuid.New(),
		Email: email,
		Phone: phone,
		FullName: full_name,
		PasswordHash: password_hash,
		Role: Role(role),
		Status: StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsDeleted: false,
	}
}