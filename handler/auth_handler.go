package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hackdaemon2/instashop/model"
	"github.com/hackdaemon2/instashop/repository"
	"github.com/hackdaemon2/instashop/util"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err // Return error if password hashing fails
	}
	return string(hashedPassword), nil
}

func newUserFromSignupRequest(signupRequest SignupRequest) (*model.User, error) {
	password, err := hashPassword(signupRequest.Password)
	if err != nil {
		return nil, err
	}

	// Map the SignupRequest to a User model
	user := &model.User{
		Email:     signupRequest.Email,
		Password:  password,
		FirstName: signupRequest.FirstName,
		LastName:  signupRequest.LastName,
		Currency:  signupRequest.UserCurrency,
		Role:      model.UserRole, // Default role
	}

	return user, nil
}

// Signup function for user registration
// @Summary Register a new user
// @Description Registers a user with the provided details
// @Tags Authentication
// @Produce		json
// @Param signup body SignupRequest true "Signup Request"
// @Success 201 {object} map[string]any{message=string, order=model.User} "User successfully registered"
// @Failure 400 {object} map[string]any{error=bool, message=string} "Invalid input"
// @Failure 500 {object} map[string]any{error=bool, message=string} "Server error"
// @Router /api/v1/user/signup [post]
func Signup(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var signupRequest SignupRequest
		if err := ctx.ShouldBindJSON(&signupRequest); err != nil { // Bind JSON and validate input
			validationError := util.ExtractValidationErrorMessage(err, signupRequest)
			util.LogAndHandleResponse(ctx, http.StatusBadRequest, ErrorResponse{Error: true, Message: validationError[0]})
			return
		}

		// Log the request data
		util.LogIncomingRequest(signupRequest)

		// validate passwords
		if signupRequest.ConfirmPassword != signupRequest.Password {
			errorResponse := ErrorResponse{Error: true, Message: "'password' and 'confirm_password' do not match"}
			util.LogAndHandleResponse(ctx, http.StatusBadRequest, errorResponse)
			return
		}

		user, err := newUserFromSignupRequest(signupRequest)
		if err != nil {
			util.LogAndHandleResponse(ctx, http.StatusInternalServerError, ErrorResponse{Error: true, Message: err.Error()})
			return
		}

		// Register user
		user, err = repository.RegisterUser(db, user)
		if err != nil {
			util.LogAndHandleResponse(ctx, http.StatusInternalServerError, ErrorResponse{Error: true, Message: err.Error()})
			return
		}

		// Send successful response
		util.LogAndHandleResponse(ctx, http.StatusCreated, UserResponse{User: user, Message: "User successfully registered"})
	}
}

// Login function for user authentication
// @Summary Authenticate a user
// @Description Authenticates a user using their email and password
// @Tags Authentication
// @Produce		json
// @Param login body LoginRequest true "Login Request"
// @Success 200 {object} map[string]any{token=string, issuer=string, issued=string, expires=string, user_id=string} "Successful authentication"
// @Failure 400 {object} map[string]any{error=bool, message=string} "Invalid input"
// @Failure 401 {object} map[string]any{error=bool, message=string} "Invalid credentials"
// @Router /api/v1/user/login [post]
func Login(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var loginRequest LoginRequest
		if err := ctx.ShouldBindJSON(&loginRequest); err != nil { // Bind JSON and validate input
			validationError := util.ExtractValidationErrorMessage(err, loginRequest)
			fmt.Println(validationError)
			response := ErrorResponse{Error: true, Message: validationError[0]}
			util.LogAndHandleResponse(ctx, http.StatusBadRequest, response)
			return
		}

		// Log the request data
		util.LogIncomingRequest(loginRequest)

		// Authenticate user
		response, err := repository.LoginUser(db, loginRequest.Email, loginRequest.Password)
		if err != nil {
			response := ErrorResponse{Error: true, Message: "Invalid credentials"}
			util.LogAndHandleResponse(ctx, http.StatusUnauthorized, response)
			return
		}

		// Send successful response
		util.LogAndHandleResponse(ctx, http.StatusOK, response)
	}
}
