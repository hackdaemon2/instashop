package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

const (
	EMAIL            = "test@test.com"
	MOCK_ERROR       = "Error creating mock db: %v"
	OPEN_ERROR       = "Error opening gorm connection: %v"
	SIGNUP           = "/signup"
	CONTEXT_TYPE     = "Content-Type"
	APPLICATION_JSON = "application/json"
	LOGIN            = "/login"
	SELECT_QUERY     = "SELECT * FROM `users` WHERE (email = ? AND is_deleted = false)"
)

func createSignupRequest() SignupRequest {
	return SignupRequest{
		Email:           EMAIL,
		Password:        "password123",
		ConfirmPassword: "password123",
		UserCurrency:    "NGN",
		FirstName:       "John",
		LastName:        "Doe",
	}
}

func createLoginRequest() LoginRequest {
	return LoginRequest{
		Email:    EMAIL,
		Password: "mayfay_2018@M1",
	}
}

// Helper function to generate JSON request body
func generateRequestBody(t *testing.T, body any) *bytes.Buffer {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Error encoding JSON: %v", err)
	}
	return bytes.NewBuffer(jsonBody)
}

// Helper function to create a test context
func createTestContext(reqBody any, endpoint string, t *testing.T) (*httptest.ResponseRecorder, *gin.Context) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", endpoint, generateRequestBody(t, reqBody))
	c.Request.Header.Set(CONTEXT_TYPE, APPLICATION_JSON)
	return w, c
}

// Valid signup request with matching passwords returns 201 and success message
func TestSignupHandlerSuccessfulRegistration(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("MOCK_ERROR: %v", err)
	}
	defer db.Close()

	gdb, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("OPEN_ERROR: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE (email = ? AND is_deleted = false) ORDER BY `users`.`id` ASC LIMIT 1")).
		WithArgs(EMAIL).
		WillReturnError(gorm.ErrRecordNotFound)

	sql := "INSERT INTO `users` (`email`,`password`,`first_name`,`last_name`,`user_currency`,`user_guid`,`role`,`created_at`,`updated_at`) VALUES (?,?,?,?,?,?,?,?,?)"
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(sql)).
		WithArgs(EMAIL, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `is_deleted` FROM `users` WHERE (id = ?)")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"is_deleted"}).AddRow(0))
	mock.ExpectCommit()

	reqBody := createSignupRequest()

	w, c := createTestContext(reqBody, SIGNUP, t)
	Signup(gdb)(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "User successfully registered")
}

// Request JSON successfully binds to SignupRequest struct
func TestSignupHandlerBindsJSON(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf(MOCK_ERROR, err)
	}
	defer db.Close()

	_, err = gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf(OPEN_ERROR, err)
	}

	reqBody := createSignupRequest()

	_, c := createTestContext(reqBody, SIGNUP, t)

	var captured SignupRequest
	c.ShouldBindJSON(&captured)

	assert.Equal(t, reqBody.Email, captured.Email)
	assert.Equal(t, reqBody.Password, captured.Password)
	assert.Equal(t, reqBody.ConfirmPassword, captured.ConfirmPassword)
}

// Request JSON successfully binds to LoginRequest struct
func TestLoginHandlerBindsJSON(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf(MOCK_ERROR, err)
	}
	defer db.Close()

	_, err = gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf(OPEN_ERROR, err)
	}

	reqBody := createLoginRequest()

	_, c := createTestContext(reqBody, SIGNUP, t)

	var captured SignupRequest
	c.ShouldBindJSON(&captured)

	assert.Equal(t, reqBody.Email, captured.Email)
	assert.Equal(t, reqBody.Password, captured.Password)
}

// Invalid JSON format in request body returns 400
func TestSignupHandlerInvalidJSON(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf(MOCK_ERROR, err)
	}
	defer db.Close()

	gdb, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf(OPEN_ERROR, err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	invalidJSON := `{"email": EMAIL, "password": "pass123", invalid json}`
	c.Request = httptest.NewRequest("POST", SIGNUP, strings.NewReader(invalidJSON))
	c.Request.Header.Set(CONTEXT_TYPE, APPLICATION_JSON)

	Signup(gdb)(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Invalid JSON format in request body returns 400
func TestLoginHandlerInvalidJSON(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf(MOCK_ERROR, err)
	}
	defer db.Close()

	gdb, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf(OPEN_ERROR, err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	invalidJSON := `{"email": EMAIL, "password": invalid json}`
	c.Request = httptest.NewRequest("POST", LOGIN, strings.NewReader(invalidJSON))
	c.Request.Header.Set(CONTEXT_TYPE, APPLICATION_JSON)

	Signup(gdb)(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Missing required fields in request returns validation error
func TestSignupHandlerMissingRequiredFields(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf(MOCK_ERROR, err)
	}
	defer db.Close()

	gdb, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf(OPEN_ERROR, err)
	}

	reqBody := SignupRequest{
		Email:           "",
		Password:        "",
		ConfirmPassword: "",
		UserCurrency:    "",
		FirstName:       "",
		LastName:        "",
	}

	w, c := createTestContext(reqBody, SIGNUP, t)
	Signup(gdb)(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "is required")
}

// Missing required fields in request returns validation error
func TestLoginHandlerMissingRequiredFields(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf(MOCK_ERROR, err)
	}
	defer db.Close()

	gdb, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf(OPEN_ERROR, err)
	}

	reqBody := LoginRequest{
		Email:    "",
		Password: "",
	}

	w, c := createTestContext(reqBody, SIGNUP, t)
	Login(gdb)(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "is required")
}

// Password and confirm password mismatch returns 400
func TestSignupHandlerPasswordMismatch(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf(MOCK_ERROR, err)
	}
	defer db.Close()

	gdb, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf(OPEN_ERROR, err)
	}

	reqBody := SignupRequest{
		Email:           EMAIL,
		Password:        "password123",
		ConfirmPassword: "wrongpassword",
		UserCurrency:    "NGN",
		FirstName:       "John",
		LastName:        "Doe",
	}

	w, c := createTestContext(reqBody, SIGNUP, t)
	Signup(gdb)(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "do not match")
}

// Duplicate email returns 400
func TestSignupHandlerDuplicateEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(MOCK_ERROR, err)
	}
	defer db.Close()

	gdb, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf(OPEN_ERROR, err)
	}

	// Use regex to match the core part of the query
	mock.ExpectQuery(regexp.QuoteMeta(SELECT_QUERY)).
		WithArgs(EMAIL).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password"}).
			AddRow("1", EMAIL, "$2y$10$XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"))

	reqBody := createSignupRequest()

	w, c := createTestContext(reqBody, SIGNUP, t)
	Signup(gdb)(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Regexp(t, `.*user with email test@test.com already exists.*`, w.Body.String())
}

// Successful Login
func TestLoginSuccessful(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(MOCK_ERROR, err)
	}
	defer db.Close()

	gdb, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf(OPEN_ERROR, err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(SELECT_QUERY)).
		WithArgs(EMAIL).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password"}).
			AddRow("1", EMAIL, "$2a$10$BvynkDL3zqY9wn8J6QFUD.XiSETqPtPPvs5VjH//EAJflvXNP3wRe"))

	reqBody := createLoginRequest()

	w, c := createTestContext(reqBody, LOGIN, t)
	Login(gdb)(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "token")
}
