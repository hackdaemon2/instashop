package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/hackdaemon2/instashop/model"
)

const (
	TEST_USER_ID         = "test_user_id"
	TEST_ORDER_REF       = "test_order_ref"
	TEST_PRODUCT_CODE    = "product123"
	TEST_CURRENCY        = "USD"
	ORDER_ENDPOINT       = "/api/v1/user/order"
	SELECT_ORDER_QUERY   = "SELECT * FROM `orders` WHERE (order_reference = ? AND is_deleted = false)"
	SELECT_USER_QUERY    = "SELECT * FROM `users` WHERE (user_guid = ? AND is_deleted = false) ORDER BY `users`.`id` ASC LIMIT 1"
	SELECT_PRODUCT_QUERY = "SELECT * FROM `products` INNER JOIN `order_products` ON `order_products`.`product_id` = `products`.`id` WHERE (`order_products`.`order_id` IN (?)) AND (is_deleted = ?)"
)

func createOrderRequest() OrderRequest {
	return OrderRequest{
		UserID:         TEST_USER_ID,
		OrderReference: TEST_ORDER_REF,
		Products: []ProductDTO{
			{
				Code:     TEST_PRODUCT_CODE,
				Quantity: 1,
			},
		},
	}
}

func createUpdateOrderRequest() UpdateOrderRequest {
	return UpdateOrderRequest{
		OrderStatus: string(model.Shipped),
	}
}

func generateOrderRequestBody(t *testing.T, body any) *bytes.Buffer {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Error encoding JSON: %v", err)
	}
	return bytes.NewBuffer(jsonBody)
}

func createOrderTestContext(reqBody any, endpoint string, t *testing.T) (*httptest.ResponseRecorder, *gin.Context) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", endpoint, generateRequestBody(t, reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", TEST_USER_ID) // Set user_id for authenticated requests
	return w, c
}

// Helper function to create a mock user
func createMockUser() *model.User {
	user := &model.User{
		ID:       1,
		UserID:   TEST_USER_ID,
		Currency: TEST_CURRENCY,
	}
	return user
}

// Helper function to create a mock product
func createMockProduct() *model.Product {
	product := &model.Product{
		ID:          1,
		ProductCode: TEST_PRODUCT_CODE,
		Name:        "Test Product",
		Price:       decimal.NewFromFloat(10.00),
		Stock:       10,
		Currency:    TEST_CURRENCY,
	}
	return product
}

// Helper function to create a mock order
func createMockOrder() *model.Order {
	order := &model.Order{
		ID:             1,
		UserID:         1,
		OrderReference: TEST_ORDER_REF,
		Status:         model.Pending,
		TotalPrice:     decimal.NewFromFloat(10.00),
	}
	return order
}

// PlaceOrder: User not found
func TestPlaceOrderUserNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(MOCK_ERROR, err)
	}
	defer db.Close()

	gdb, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf(OPEN_ERROR, err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(SELECT_ORDER_QUERY)).
		WithArgs(TEST_ORDER_REF).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectQuery(regexp.QuoteMeta(SELECT_USER_QUERY)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectQuery(regexp.QuoteMeta(SELECT_PRODUCT_QUERY)).
		WithArgs(1, false).
		WillReturnError(gorm.ErrRecordNotFound)

	orderRequest := createOrderRequest()
	w, c := createTestContext(orderRequest, ORDER_ENDPOINT, t)
	PlaceOrder(gdb)(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), USER_NOT_FOUND_ERROR)
}

// PlaceOrder: Product not found
func TestPlaceOrderProductNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(MOCK_ERROR, err)
	}
	defer db.Close()

	gdb, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf(OPEN_ERROR, err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(SELECT_ORDER_QUERY)).
		WithArgs(TEST_ORDER_REF).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectQuery(regexp.QuoteMeta(SELECT_USER_QUERY)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_guid", "currency"}).AddRow(1, TEST_USER_ID, TEST_CURRENCY))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `products` WHERE (product_code = ? AND is_deleted = false) ORDER BY `products`.`id` ASC LIMIT 1")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectQuery(regexp.QuoteMeta(SELECT_PRODUCT_QUERY)).
		WithArgs(1, false).
		WillReturnError(gorm.ErrRecordNotFound)

	orderRequest := createOrderRequest()
	w, c := createTestContext(orderRequest, ORDER_ENDPOINT, t)
	PlaceOrder(gdb)(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// PlaceOrder: Order Already Exists
func TestPlaceOrderAlreadyExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(MOCK_ERROR, err)
	}
	defer db.Close()

	gdb, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf(OPEN_ERROR, err)
	}

	w := placeOrderTest(t, mock, gdb)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "Order already exists")
}

func placeOrderTest(t *testing.T, mock sqlmock.Sqlmock, gdb *gorm.DB) *httptest.ResponseRecorder {
	mock.ExpectQuery(regexp.QuoteMeta(SELECT_ORDER_QUERY)).
		WithArgs(TEST_ORDER_REF).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_guid", "currency"}).AddRow(1, TEST_USER_ID, TEST_CURRENCY))

	mock.ExpectQuery(regexp.QuoteMeta(SELECT_PRODUCT_QUERY)).
		WithArgs(1, false).
		WillReturnError(gorm.ErrRecordNotFound)

	orderRequest := createOrderRequest()
	w, c := createTestContext(orderRequest, ORDER_ENDPOINT, t)
	PlaceOrder(gdb)(c)
	return w
}
