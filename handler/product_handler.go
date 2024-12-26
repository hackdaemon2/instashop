package handler

import (
	"log"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hackdaemon2/instashop/model"
	"github.com/hackdaemon2/instashop/repository"
	"github.com/hackdaemon2/instashop/util"
	"github.com/jinzhu/gorm"
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

func roundToTwoDecimals(value float64) float64 {
	// Multiply, round, and divide to keep 2 decimal precision
	return math.Round(value*100) / 100
}

func newProduct(createNewProduct CreateProductRequest, user *model.User) *model.Product {
	return &model.Product{
		Currency:    createNewProduct.Currency,
		Name:        createNewProduct.Name,
		Description: createNewProduct.Description,
		ProductCode: uuid.New().String(),
		Price:       createNewProduct.Price,
		Stock:       createNewProduct.Stock,
		UserID:      user.ID,
	}
}

// Helper function for error handling and response
func handleProductError(ctx *gin.Context, statusCode int, message string) {
	util.LogAndHandleResponse(ctx, statusCode, util.ErrorResponse{Error: true, ErrorMessage: message})
}

// GetProduct retrieves a product by its product code
// @Summary Get a product by its product code
// @Description Retrieve product details using the product code
// @Tags Products
// @SecurityDefinitions.apiKey Bearer
// @in header
// @name Authorization
// @Produce		json
// @Security BearerAuth
// @Param Authorization header string true "Bearer Token"
// @Param product_code path string true "Product Code"
// @Success 200 {object} handler.ProductResponse{product=model.Product, message=string} "Product successfully retrieved"
// @Failure 404 {object} handler.ProductResponse{product=model.Product, message=string} "No product found"
// @Failure 500 {object} util.ErrorResponse{error=bool, error_message=string} "Failed to retrieve product"
// @Router /api/v1/product/{product_code} [get]
func GetProduct(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		productCode := ctx.Param("product_code")

		product, err := repository.GetProduct(db, productCode)
		if err != nil {
			if err.Error() == repository.PRODUCT_NOT_FOUND_ERROR {
				handleProductError(ctx, http.StatusNotFound, repository.PRODUCT_NOT_FOUND_ERROR)
				return
			}
			handleProductError(ctx, http.StatusInternalServerError, PRODUCT_RETRIEVAL_ERROR)
			return
		}

		message := "Product successfully retrieved"
		status := http.StatusOK

		if product == nil {
			message = "No product found"
			status = http.StatusNotFound
		}

		response := ProductResponse{
			Product: product,
			Message: message,
		}

		util.LogAndHandleResponse(ctx, status, response)
	}
}

// UpdateProduct updates an existing product
// @Summary Update an existing product
// @Description Update product details
// @Tags Products
// @SecurityDefinitions.apiKey Bearer
// @in header
// @name Authorization
// @Produce		json
// @Security BearerAuth
// @Param Authorization header string true "Bearer Token"
// @Param product_code path string true "Product Code"
// @Param product body UpdateProductRequest true "Product Data"
// @Success 200 {object} handler.ProductResponse{product=model.Product, message=string}
// @Failure 400 {object} util.ErrorResponse{error=bool, error_message=string}
// @Failure 404 {object} util.ErrorResponse{error=bool, error_message=string}
// @Failure 500 {object} util.ErrorResponse{error=bool, error_message=string}
// @Router /api/v1/admin/product/{product_code} [put]
func UpdateProduct(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var updateProduct UpdateProductRequest
		if err := ctx.ShouldBindJSON(&updateProduct); err != nil {
			validationError := util.ExtractValidationErrorMessage(err, updateProduct)
			handleProductError(ctx, http.StatusBadRequest, validationError[0])
			return
		}

		util.LogIncomingRequest(updateProduct)

		productCode := ctx.Param("product_code")

		product, err := repository.GetProduct(db, productCode)
		if err != nil {
			if err.Error() == repository.PRODUCT_NOT_FOUND_ERROR {
				handleProductError(ctx, http.StatusNotFound, repository.PRODUCT_NOT_FOUND_ERROR)
				return
			}
			handleProductError(ctx, http.StatusInternalServerError, PRODUCT_RETRIEVAL_ERROR)
			return
		}

		// Update product fields
		product.Description = updateProduct.Description
		product.Name = updateProduct.Name
		product.Price = updateProduct.Price
		product.Currency = updateProduct.Currency
		product.Stock = updateProduct.Stock

		updatedProduct, err := repository.UpdateProduct(db, product)
		if err != nil {
			handleProductError(ctx, http.StatusInternalServerError, "Failed to update product")
			return
		}

		response := ProductResponse{
			Product: updatedProduct,
			Message: "Product updated successfully",
		}

		util.LogAndHandleResponse(ctx, http.StatusOK, response)
	}
}

// DeleteProduct deletes a product in the database
// @Summary Delete a product
// @Description Delete a product by its product code
// @Tags Products
// @SecurityDefinitions.apiKey Bearer
// @in header
// @name Authorization
// @Produce		json
// @Security BearerAuth
// @Param Authorization header string true "Bearer Token"
// @Param product_code path string true "Product Code"
// @Success 200 {object} util.ErrorResponse{error=bool, error_message=string} "Product has been successfully deleted"
// @Failure 404 {object} util.ErrorResponse{error=bool, error_message=string} "Product not found"
// @Failure 500 {object} util.ErrorResponse{error=bool, error_message=string} "Error in deleting product"
// @Router /api/v1/admin/product/{product_code} [delete]
func DeleteProduct(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		productCode := ctx.Param("product_code")

		product, err := repository.GetProduct(db, productCode)
		if err != nil {
			if err.Error() == repository.PRODUCT_NOT_FOUND_ERROR {
				handleProductError(ctx, http.StatusNotFound, repository.PRODUCT_NOT_FOUND_ERROR)
				return
			}
			handleProductError(ctx, http.StatusInternalServerError, PRODUCT_RETRIEVAL_ERROR)
			return
		}

		err = repository.DeleteProduct(db, product)
		if err != nil {
			handleProductError(ctx, http.StatusInternalServerError, "Error in deleting product")
			return
		}

		response := util.ErrorResponse{
			Error:        false,
			ErrorMessage: "Product has been successfully deleted",
		}

		util.LogAndHandleResponse(ctx, http.StatusOK, response)
	}
}

// CreateProduct creates a new product in the database
// @Summary Create a new product
// @Description Add a new product to the database
// @Tags Products
// @SecurityDefinitions.apiKey Bearer
// @in header
// @name Authorization
// @Produce		json
// @Security BearerAuth
// @Param Authorization header string true "Bearer Token"
// @Param product body CreateProductRequest true "Product Data"
// @Success 201 {object} handler.ProductResponse{product=model.Product, message=string}
// @Failure 400 {object} util.ErrorResponse{error=bool, error_message=string}
// @Failure 500 {object} util.ErrorResponse{error=bool, error_message=string}
// @Router /api/v1/admin/product [post]
func CreateProduct(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var createProductRequest CreateProductRequest
		if err := ctx.ShouldBindJSON(&createProductRequest); err != nil {
			validationError := util.ExtractValidationErrorMessage(err, createProductRequest)
			handleProductError(ctx, http.StatusBadRequest, validationError[0])
			return
		}

		util.LogIncomingRequest(createProductRequest)

		var err error
		var user *model.User

		if authenticatedUserID, exists := ctx.Get("user_id"); exists {
			user, err = repository.FindUserBy(db, "user_guid", authenticatedUserID.(string))
			if err != nil || user == nil {
				log.Printf("error => %v\nUser is null => %v", err, user == nil)
				handleProductError(ctx, http.StatusBadRequest, INVALID_USER_INPUT)
				return
			}

			// Create the new product
			product := newProduct(createProductRequest, user)

			savedProduct, err := repository.CreateProduct(db, *product)
			if err != nil {
				handleProductError(ctx, http.StatusInternalServerError, "Failed to create product")
				return
			}

			response := ProductResponse{
				Product: savedProduct,
				Message: "Product created successfully",
			}

			util.LogAndHandleResponse(ctx, http.StatusCreated, response)
			return
		}

		handleProductError(ctx, http.StatusBadRequest, INVALID_USER_INPUT)
	}
}
