package handler

import "github.com/hackdaemon2/instashop/model"

const (
	ZERO                 = 0
	USER_NOT_FOUND_ERROR = "User not found"
	PLACE_ORDER_ERROR    = "error occured in placing order"
)

type ProductDTO struct {
	Code     string `json:"product_code" binding:"required"`
	Quantity uint   `json:"product_quantity" binding:"required"`
}

type OrderRequest struct {
	UserID         string       `json:"user_id" binding:"required"`
	OrderReference string       `json:"order_reference" binding:"required"`
	Products       []ProductDTO `json:"products" binding:"required"`
}

type UpdateOrderRequest struct {
	OrderStatus string `json:"order_status" binding:"required"`
}

type ListOrderResponse struct {
	Orders      []*model.Order `json:"orders"`
	Message     string         `json:"message"`
	TotalOrders int            `json:"total_orders"`
	TotalPages  int            `json:"total_pages"`
	Page        int            `json:"page"`
	Size        int            `json:"size"`
}

type OrderResponse struct {
	Order   *model.Order `json:"order"`
	Message string       `json:"message"`
}

type UserResponse struct {
	User    *model.User `json:"user"`
	Message string      `json:"message"`
}

type ProductResponse struct {
	Product *model.Product `json:"product"`
	Message string         `json:"message"`
}
