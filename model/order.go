package model

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

// Define a custom type for the OrderStatus
type OrderStatus string

// Define constants for each enum value
const (
	Pending   OrderStatus = "Pending"
	Shipped   OrderStatus = "Shipped"
	Delivered OrderStatus = "Delivered"
	Cancelled OrderStatus = "Cancelled"
)

type Order struct {
	ID             uint            `json:"-" gorm:"primary_key"`
	UserID         uint            `json:"-" gorm:"column:user_id;index"`
	User           User            `json:"-" gorm:"foreignKey:UserID"`                                // Establish the relationship with User
	Status         OrderStatus     `json:"order_status" gorm:"column:order_status" example:"Pending"` // Pending, Shipped, Delivered, Canceled
	TotalPrice     decimal.Decimal `json:"total_price" gorm:"column:total_price;type:decimal(10,2)" example:"10.50"`
	OrderReference string          `json:"order_reference" gorm:"column:order_reference;index" example:"order123"`
	IsDeleted      bool            `json:"-" gorm:"column:is_deleted;default:false"`
	Products       []Product       `json:"products" gorm:"many2many:order_products;"`
	CreatedAt      time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt      time.Time       `json:"updated_at" gorm:"column:updated_at"`
}

func (order *Order) BeforeCreate(tx *gorm.DB) (err error) {
	// You can modify the data before inserting it into the DB
	now := time.Now()
	order.CreatedAt = now
	order.UpdatedAt = now
	return nil
}

func (order *Order) BeforeUpdate(tx *gorm.DB) (err error) {
	now := time.Now()
	order.UpdatedAt = now
	return nil
}
