package repository

import (
	"github.com/hackdaemon2/instashop/model"
	"github.com/jinzhu/gorm"
)

const IS_DELETED_CLAUSE = "is_deleted = ?"

func FindOrder(db *gorm.DB, orderReference string) (*model.Order, error) {
	var order model.Order
	query := "order_reference = ? AND is_deleted = false"
	err := db.Where(query, orderReference).Preload("Products", IS_DELETED_CLAUSE, false).Find(&order).Error
	return &order, err
}

func CreateOrder(db *gorm.DB, order model.Order) (*model.Order, error) {
	if err := db.Create(&order).Error; err != nil { // Create the order
		return nil, err
	}
	return &order, nil
}

func UpdateOrder(db *gorm.DB, order model.Order) (*model.Order, error) {
	// Update order fields
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Model(&order).Preload("Products", IS_DELETED_CLAUSE, false).Save(&order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	err := tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return &order, nil
	}

	return nil, err
}

func GetUserOrder(db *gorm.DB, userID, orderReference string) (*model.Order, error) {
	var orders model.Order
	query := "user_id = ? AND order_reference = ? AND is_deleted = false"
	err := db.Where(query, userID, orderReference).Preload("Products", IS_DELETED_CLAUSE, false).Find(&orders).Error
	return &orders, err
}

func GetUserOrders(db *gorm.DB, userID, orderStatus string, page, limit int) ([]*model.Order, int, error) {
	var orders []*model.Order
	var totalOrders int

	query := db.Where("user_id = ? AND is_deleted = false", userID)

	if orderStatus != "" {
		query = query.Where("order_status = ?", orderStatus)
	}

	err := query.Model(&model.Order{}).Count(&totalOrders).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = query.Preload("Products", IS_DELETED_CLAUSE, false).Limit(limit).Offset(offset).Find(&orders).Error

	return orders, totalOrders, err
}
