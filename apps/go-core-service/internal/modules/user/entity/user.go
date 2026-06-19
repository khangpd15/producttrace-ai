package entity

import (
	"time"

	"github.com/google/uuid"
)

type Role string
type Status string

const (
	RoleAdmin    Role = "ADMIN"
	RoleCustomer Role = "CUSTOMER"
	RoleStaff    Role = "STAFF"
	RoleDealer   Role = "DEALER"
)

const (
	StatusActive    Status = "ACTIVE"
	StatusBanned    Status = "BANNED"
	StatusSuspended Status = "SUSPENDED"
	StatusPending   Status = "PENDING"
)

type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Email        string    `json:"email" gorm:"unique;not null"`
	Phone        string    `json:"phone" gorm:"unique"`
	FullName     string    `json:"full_name"`
	PasswordHash string    `json:"-" gorm:"column:password_hash"`
	Role         Role      `json:"role" gorm:"default:CUSTOMER"`
	Status       Status    `json:"status" gorm:"default:ACTIVE"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsDeleted    bool      `json:"is_deleted" gorm:"default:false"`
}

func NewUser(email, phone, fullName, passwordHash, role string) *User {
	if role == "" {
		role = string(RoleCustomer)
	}
	return &User{
		ID:           uuid.New(),
		Email:        email,
		Phone:        phone,
		FullName:     fullName,
		PasswordHash: passwordHash,
		Role:         Role(role),
		Status:       StatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsDeleted:    false,
	}
}
