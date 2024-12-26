package handler

type ErrorResponse struct {
	Error        bool   `json:"error"`
	ErrorMessage string `json:"error_message"`
}
