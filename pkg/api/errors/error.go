package errors

// ErrorCode represents an error code as a string.
type ErrorCode string

// Error codes constants, unexported.
const (
	Unauthorized         ErrorCode = "UNAUTHORIZED"
	Error                ErrorCode = "API_ERROR" // Generic API error
	NotAllowed           ErrorCode = "NOT_ALLOWED"
	NotFound             ErrorCode = "NOT_FOUND"
	AlreadyExists        ErrorCode = "ALREADY_EXISTS"
	InternalServerError  ErrorCode = "INTERNAL_SERVER_ERROR"
	UnsupportedGrantType ErrorCode = "UNSUPPORTED_GRANT_TYPE"
)

// Represents a standardized error response for the API
type ApiError struct {
	// Code represents the HTTP status code
	Code int `json:"code"`

	// Error represents a predefined error code from the enum package
	Error ErrorCode `json:"error"`

	// Details contains additional error information (optional)
	Details interface{} `json:"details,omitempty"`
}
