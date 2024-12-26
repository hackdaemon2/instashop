package handler

import (
	"github.com/shopspring/decimal"
)

const (
	PRODUCT_RETRIEVAL_ERROR = "Failed to retrieve product"
	INVALID_USER_INPUT      = "Invalid input"
)

type ProductCommonData struct {
	Description string          `json:"product_description"`
	Name        string          `json:"product_name" binding:"required,min=3"`
	Price       decimal.Decimal `json:"price" binding:"required"`
	Stock       uint            `json:"stock" binding:"required,numeric"`
	Currency    string          `json:"currency" binding:"required,min=3,max=3"`
}

type CreateProductRequest struct {
	ProductCommonData
	UserID string `json:"user_id" binding:"required"`
}

type UpdateProductRequest struct {
	ProductCommonData
}
