package handler

type SignupRequest struct {
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8"`
	UserCurrency    string `json:"user_currency" binding:"required,min=3,max=3"`
	ConfirmPassword string `json:"confirm_password" binding:"required,min=8"`
	FirstName       string `json:"first_name" binding:"required"`
	LastName        string `json:"last_name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
