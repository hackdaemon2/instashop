package model

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

type Product struct {
	ID          uint            `json:"-" gorm:"primary_key"`
	Name        string          `json:"product_name" gorm:"column:product_name"`
	Description string          `json:"product_description" gorm:"column:product_description;not null;size:255"`
	ProductCode string          `json:"product_code" gorm:"column:product_code;unique;not null;size:255"`
	Price       decimal.Decimal `json:"price" gorm:"column:price;type:decimal(10,2);not null;default:0"`
	Stock       uint            `json:"stock" gorm:"column:stock"`
	IsDeleted   bool            `json:"-" gorm:"column:is_deleted;default:false"`
	Currency    string          `json:"currency" gorm:"column:currency;not null;size:3"`
	UserID      uint            `json:"-" gorm:"column:user_id"`    // Foreign key for User
	User        User            `json:"-" gorm:"foreignKey:UserID"` // Establish the relationship with User
	CreatedAt   time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time       `json:"updated_at" gorm:"column:updated_at"`
}

func (product *Product) BeforeCreate(tx *gorm.DB) (err error) {
	// You can modify the data before inserting it into the DB
	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now
	return nil
}

func (product *Product) BeforeUpdate(tx *gorm.DB) (err error) {
	now := time.Now()
	product.UpdatedAt = now
	return nil
}
