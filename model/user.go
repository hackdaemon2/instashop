package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type Role string

const (
	AdminRole = "admin"
	UserRole  = "user"
)

type User struct {
	ID        uint      `json:"-" gorm:"primary_key"`
	Email     string    `gorm:"column:email;unique"`
	Password  string    `json:"-" gorm:"column:password;not null"`
	FirstName string    `json:"first_name" gorm:"column:first_name;not null;size:255"`
	LastName  string    `json:"last_name" gorm:"column:last_name;not null;size:255"`
	IsDeleted bool      `json:"-" gorm:"column:is_deleted;default:false"`
	Currency  string    `json:"user_currency" gorm:"column:user_currency;not null;size:3"`
	UserID    string    `json:"user_id" gorm:"column:user_guid;not null;unique"`
	Role      Role      `json:"user_role" gorm:"column:role;not null"` // user or admin
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (user *User) BeforeCreate(tx *gorm.DB) (err error) {
	// You can modify the data before inserting it into the DB
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.UserID = uuid.New().String()
	user.IsDeleted = false
	return nil
}

func (user *User) BeforeUpdate(tx *gorm.DB) (err error) {
	now := time.Now()
	user.UpdatedAt = now
	return nil
}
