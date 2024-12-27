package repository

import (
	"errors"
	"fmt"

	"github.com/hackdaemon2/instashop/model"
	"github.com/jinzhu/gorm"
)

// Helper function to find a product by its product code
func findByProductCode(db *gorm.DB, productCode string) (*model.Product, error) {
	var product model.Product
	if err := db.Where("product_code = ? AND is_deleted = false", productCode).First(&product).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.New(PRODUCT_NOT_FOUND_ERROR) // Not found
		}
		return nil, err // Other errors
	}
	return &product, nil
}

// GetProduct retrieves a product by its product code
func GetProduct(db *gorm.DB, productCode string) (*model.Product, error) {
	product, err := findByProductCode(db, productCode)
	if err != nil {
		return nil, err
	}
	return product, nil
}

// UpdateProduct updates an existing product
func UpdateProduct(db *gorm.DB, product *model.Product) (*model.Product, error) {
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// Update product fields
	if err := tx.Model(product).Save(product).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	err := tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return product, nil
	}

	return nil, err
}

func DeleteProduct(db *gorm.DB, product *model.Product) error {
	if err := db.Model(product).Update("is_deleted", true).Error; err != nil {
		fmt.Printf("error: %v", err)
		return err
	}
	return nil
}

// CreateProduct creates a new product in the database
func CreateProduct(db *gorm.DB, product model.Product) (*model.Product, error) {
	if err := db.Create(&product).Error; err != nil { // Create the product
		return nil, err
	}
	return &product, nil
}
