package errors

import "github.com/gin-gonic/gin"

// ErrorCode represents an error code as a string.
type ErrorCode string

// Error codes constants, unexported.
const (
	Unauthorized            ErrorCode = "UNAUTHORIZED"
	Error                   ErrorCode = "API_ERROR" // Generic API error
	NotAllowed              ErrorCode = "NOT_ALLOWED"
	NotFound                ErrorCode = "NOT_FOUND"
	AlreadyExists           ErrorCode = "ALREADY_EXISTS"
	InternalServerError     ErrorCode = "INTERNAL_SERVER_ERROR"
	UnsupportedGrantType    ErrorCode = "UNSUPPORTED_GRANT_TYPE"
	InvalidSessionToken     ErrorCode = "INVALID_SESSION_TOKEN"
	MissingClientID         ErrorCode = "MISSING_CLIENT_ID"
	InvalidClientID         ErrorCode = "INVALID_CLIENT_ID"
	MissingResponseType     ErrorCode = "MISSING_RESPONSE_TYPE"
	UnsupportedResponseType ErrorCode = "UNSUPPORTED_RESPONSE_TYPE"
	MissingCodeChallenge    ErrorCode = "MISSING_CODE_CHALLENGE"
	MissingState            ErrorCode = "MISSING_STATE"
	InvalidState            ErrorCode = "INVALID_STATE"
	MissingRedirectURI      ErrorCode = "MISSING_REDIRECT_URI"
	InvalidRedirectURI      ErrorCode = "INVALID_REDIRECT_URI"
)

// Represents a standardized error response for the API
type ApiError struct {
	// Code represents the HTTP status code
	Code int `json:"code"`

	// Error represents a predefined error code from the enum package
	Error ErrorCode `json:"error"`

	// Details contains additional error information (optional)
	Details any `json:"details,omitempty"`
}

func SendErrorResponse(c *gin.Context, httpCode int, code ErrorCode, details any) {
	c.JSON(httpCode, ApiError{
		Code:    httpCode,
		Error:   code,
		Details: details,
	})
}
