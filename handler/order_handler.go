package handler

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hackdaemon2/instashop/model"
	"github.com/hackdaemon2/instashop/repository"
	"github.com/hackdaemon2/instashop/util"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

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

func isNotRecordNotFoundError(err error) bool {
	return err != nil && !gorm.IsRecordNotFoundError(err)
}

// newOrder is a helper function to create a new order model
func newOrder(userID uint, products []model.Product, orderReference string, totalPrice decimal.Decimal) model.Order {
	return model.Order{
		UserID:         userID,
		OrderReference: orderReference,
		Status:         model.Pending,
		TotalPrice:     totalPrice,
		Products:       products,
	}
}

// isCurrencyMismatch is a function to check for currency mismatch across products
func isCurrencyMismatch(products []model.Product, expectedCurrency string) bool {
	for _, product := range products {
		if product.Currency != expectedCurrency {
			return true
		}
	}
	return false
}

// isValidOrderStatus is a function to validate order status
func isValidOrderStatus(status string) bool {
	validStatuses := map[model.OrderStatus]bool{
		model.Pending:   true,
		model.Shipped:   true,
		model.Delivered: true,
		model.Cancelled: true,
	}
	return validStatuses[model.OrderStatus(status)]
}

// handleOrderError is a function to handle errors consistently in all
// the functions where an err might occur
func handleOrderError(ctx *gin.Context, statusCode int, message string, err error) {
	if err != nil {
		log.Println(err.Error())
	}

	errorResponse := util.ErrorResponse{
		Error:        true,
		ErrorMessage: message,
	}

	util.LogAndHandleResponse(ctx, statusCode, errorResponse)
}

// Validate user existence
func validateUser(db *gorm.DB, userID string) (*model.User, error) {
	return repository.FindUserBy(db, "user_guid", userID)
}

// Validate products and calculate total price
func validateProducts(db *gorm.DB, productsDTO []ProductDTO, user *model.User) ([]model.Product, decimal.Decimal, error) {
	var products []model.Product

	totalPrice := decimal.NewFromFloat(0)
	zero := decimal.NewFromFloat(0)

	for _, productDTO := range productsDTO {
		product, err := repository.GetProduct(db, productDTO.Code)
		if err != nil {
			return nil, zero, fmt.Errorf("Product with code %s is not found", productDTO.Code)
		}

		// check if the quantity is a valid value
		if productDTO.Quantity <= ZERO {
			return nil, zero, fmt.Errorf("Invalid quantity for product %s (code: %s)", product.Name, product.ProductCode)
		}

		// check if the product
		if product.Stock == ZERO {
			return nil, zero, fmt.Errorf("Product %s is out of stock", productDTO.Code)
		}

		if product.Stock < productDTO.Quantity {
			return nil, zero, fmt.Errorf("Product: %s is not enough in stock. There are only %d left", product.Name, product.Stock)
		}

		quantity := decimal.NewFromInt(int64(productDTO.Quantity))
		totalPrice = totalPrice.Add(quantity.Mul(product.Price))
		product.Stock = product.Stock - productDTO.Quantity

		_, err = repository.UpdateProduct(db, product) // update product stock
		if err != nil {
			return nil, zero, err
		}

		products = append(products, *product)
	}

	if isCurrencyMismatch(products, user.Currency) {
		return nil, zero, fmt.Errorf("Product currency mismatch with user's currency: %s", user.Currency)
	}

	return products, totalPrice, nil
}

// CancelUserOrder cancels the specific order placed by a user
// @Summary Cancel a user order
// @Description Cancels an order associated with the provided order reference
// @Tags Orders
// @SecurityDefinitions.apiKey Bearer
// @in header
// @name Authorization
// @Produce		json
// @Security BearerAuth
// @Param Authorization header string true "Bearer Token"
// @Param user_id header string true "Authenticated User ID"
// @Param order_reference path string true "Order Reference"
// @Success 200 {object} handler.OrderResponse{message=string, order=model.Order} "Order cancelled successfully"
// @Failure 400 {object} util.ErrorResponse{error=bool, error_message=string} "Invalid input or Order in %s status cannot be cancelled"
// @Failure 401 {object} util.ErrorResponse{error=bool, error_message=string} "Unauthorized access"
// @Failure 404 {object} util.ErrorResponse{error=bool, error_message=string} "Order not found"
// @Failure 500 {object} util.ErrorResponse{error=bool, error_message=string} "Failed to update order"
// @Router /api/v1/user/order/{order_reference}/cancel [PUT]
func CancelUserOrder(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		orderReference := ctx.Param("order_reference")
		authUserID, exists := ctx.Get("user_id")
		if !exists {
			handleOrderError(ctx, http.StatusUnauthorized, "Unauthorized access", nil)
			return
		}

		user, err := validateUser(db, authUserID.(string))
		if err != nil || user == nil {
			if err != nil {
				handleOrderError(ctx, http.StatusInternalServerError, "Error in order", err)
				return
			}
			handleOrderError(ctx, http.StatusNotFound, USER_NOT_FOUND_ERROR, err)
			return
		}

		id := strconv.Itoa(int(user.ID))

		order, err := repository.GetUserOrder(db, id, orderReference)
		if err != nil {
			handleOrderError(ctx, http.StatusNotFound, "Order not found", err)
			return
		}

		fmt.Println(order)

		if order.Status != model.Pending {
			fmt.Println(order.Status)
			errorMessage := fmt.Sprintf("Order in %s status cannot be cancelled", string(order.Status))
			handleOrderError(ctx, http.StatusBadRequest, errorMessage, err)
			return
		}

		order.Status = model.Cancelled
		updatedOrder, err := repository.UpdateOrder(db, *order)
		if err != nil {
			handleOrderError(ctx, http.StatusInternalServerError, "Failed to update order", err)
			return
		}

		response := OrderResponse{
			Order:   updatedOrder,
			Message: "Order cancelled successfully",
		}

		util.LogAndHandleResponse(ctx, http.StatusOK, response)
	}
}

// PlaceOrder godoc
// @Summary Place a new order
// @Description Creates a new order for a user with a list of products
// @Tags Orders
// @SecurityDefinitions.apiKey Bearer
// @in header
// @name Authorization
// @Produce		json
// @Security BearerAuth
// @Param Authorization header string true "Bearer Token"
// @Param order body OrderRequest true "Order Request"
// @Success 201 {object} handler.OrderResponse{message=string, order=model.Order} "Order placed successfully"
// @Failure 400 {object} util.ErrorResponse{error=bool, error_message=string} "Invalid input"
// @Failure 404 {object} util.ErrorResponse{error=bool, error_message=string} "User not found"
// @Failure 500 {object} util.ErrorResponse{error=bool, error_message=string} "Failed to place order"
// @Router /api/v1/user/order [post]
func PlaceOrder(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var orderRequest OrderRequest
		if err := ctx.ShouldBindJSON(&orderRequest); err != nil {
			validationError := util.ExtractValidationErrorMessage(err, orderRequest)
			handleOrderError(ctx, http.StatusBadRequest, validationError[0], err)
			return
		}

		existingOrder, err := repository.FindOrder(db, orderRequest.OrderReference)
		if existingOrder.ID != 0 {
			handleOrderError(ctx, http.StatusConflict, "Order already exists", err)
			return
		}

		if !gorm.IsRecordNotFoundError(err) {
			handleOrderError(ctx, http.StatusInternalServerError, PLACE_ORDER_ERROR, err)
			return
		}

		user, err := validateUser(db, orderRequest.UserID)
		if isNotRecordNotFoundError(err) {
			handleOrderError(ctx, http.StatusInternalServerError, PLACE_ORDER_ERROR, err)
			return
		}

		if user == nil || gorm.IsRecordNotFoundError(err) {
			handleOrderError(ctx, http.StatusNotFound, USER_NOT_FOUND_ERROR, err)
			return
		}

		products, totalPrice, err := validateProducts(db, orderRequest.Products, user)
		if err != nil {
			handleOrderError(ctx, http.StatusBadRequest, err.Error(), err)
			return
		}

		order := newOrder(user.ID, products, orderRequest.OrderReference, totalPrice)

		savedOrder, err := repository.CreateOrder(db, order)
		if err != nil {
			handleOrderError(ctx, http.StatusInternalServerError, "Failed to place order", err)
			return
		}

		response := OrderResponse{
			Order:   savedOrder,
			Message: "Order placed successfully",
		}

		util.LogAndHandleResponse(ctx, http.StatusCreated, response)
	}
}

// GetUserOrders godoc
// @Summary Get user orders
// @Description Retrieves all orders for a given user with an optional status filter
// @Tags Orders
// @SecurityDefinitions.apiKey Bearer
// @in header
// @name Authorization
// @Produce		json
// @Security BearerAuth
// @Param Authorization header string true "Bearer Token"
// @Param user_id query string true "User ID"
// @Param order_status query string false "Order Status (Pending, Shipped, Delivered, Cancelled)"
// @Param page query string false "Page (Default 1)"
// @Param size query string false "Size (Default 10)"
// @Success 200 {object} handler.ListOrderResponse{orders=[]model.Order, message=string, total_orders=int, total_pages=int, page=int, size=int} "List of user orders"
// @Failure 500 {object} util.ErrorResponse{error=bool, error_message=string} "Failed to retrieve orders"
// @Router /api/v1/user/order [get]
func GetUserOrders(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.Query("user_id")
		orderStatus := ctx.Query("order_status")

		page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
		if err != nil || page <= 0 {
			page = 1
		}

		limit, err := strconv.Atoi(ctx.DefaultQuery("size", "10"))
		if err != nil || limit <= 0 {
			limit = 10
		}

		user, err := validateUser(db, userID)
		if err != nil {
			handleOrderError(ctx, http.StatusInternalServerError, "error in retrieving order", err)
			return
		}

		if user == nil {
			handleOrderError(ctx, http.StatusNotFound, USER_NOT_FOUND_ERROR, err)
			return
		}

		id := strconv.Itoa(int(user.ID))

		orders, totalOrders, err := repository.GetUserOrders(db, id, orderStatus, page, limit)
		if err != nil {
			handleOrderError(ctx, http.StatusInternalServerError, "Failed to retrieve orders", err)
			return
		}

		message := "Order retrieved successfully"
		if len(orders) == 0 {
			message = "No orders found"
		}

		totalPages := int(math.Ceil(float64(totalOrders) / float64(limit)))

		response := ListOrderResponse{
			TotalOrders: totalOrders,
			Page:        page,
			TotalPages:  totalPages,
			Size:        limit,
			Message:     message,
			Orders:      orders,
		}

		util.LogAndHandleResponse(ctx, http.StatusOK, response)
	}
}

// UpdateOrderStatus Update order status
// @Summary Update order status
// @Description Updates the status of a specific order for a user
// @Tags Orders
// @SecurityDefinitions.apiKey Bearer
// @in header
// @name Authorization
// @Produce		json
// @Security BearerAuth
// @Param Authorization header string true "Bearer Token"
// @Param order_reference path string true "Order Reference"
//
//	@Param updateOrder body UpdateOrderRequest true "Update Status Request"
//
// @Success 200 {object} handler.OrderResponse{message=string, order=model.Order} "Order status updated successfully"
// @Failure 400 {object} util.ErrorResponse{error=bool, error_message=string} "Order has been (Shipped | Delivered)"
// @Router /api/v1/admin/order/{order_reference}/status [put]
func UpdateOrderStatus(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var updateRequest UpdateOrderRequest
		if err := ctx.ShouldBindJSON(&updateRequest); err != nil {
			validationError := util.ExtractValidationErrorMessage(err, updateRequest)
			handleOrderError(ctx, http.StatusBadRequest, validationError[0], err)
			return
		}

		orderReference := ctx.Param("order_reference")
		if !isValidOrderStatus(updateRequest.OrderStatus) {
			handleOrderError(ctx, http.StatusBadRequest, "Invalid order status", nil)
			return
		}

		order, err := repository.FindOrder(db, orderReference)
		if err != nil {
			handleOrderError(ctx, http.StatusNotFound, "Order not found", err)
			return
		}

		if order.Status != model.Pending && order.Status != model.Shipped {
			handleOrderError(ctx, http.StatusBadRequest, fmt.Sprintf("Order has been %s", string(order.Status)), err)
			return
		}

		order.Status = model.OrderStatus(updateRequest.OrderStatus)
		updatedOrder, err := repository.UpdateOrder(db, *order)
		if err != nil {
			handleOrderError(ctx, http.StatusInternalServerError, "Failed to update order", err)
			return
		}

		response := OrderResponse{
			Order:   updatedOrder,
			Message: "Order status updated successfully",
		}

		util.LogAndHandleResponse(ctx, http.StatusOK, response)
	}
}
