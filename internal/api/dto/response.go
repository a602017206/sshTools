package dto

// SuccessResponse is the standard success response format
type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse is the standard error response format
type ErrorResponse struct {
	Error string `json:"error"`
}

// NewSuccessResponse creates a new success response with data
func NewSuccessResponse(data interface{}) SuccessResponse {
	return SuccessResponse{Data: data}
}

// NewSuccessMessageResponse creates a new success response with a message
func NewSuccessMessageResponse(message string) SuccessResponse {
	return SuccessResponse{Message: message}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err error) ErrorResponse {
	return ErrorResponse{Error: err.Error()}
}

// NewErrorMessageResponse creates a new error response with a message string
func NewErrorMessageResponse(message string) ErrorResponse {
	return ErrorResponse{Error: message}
}
