// Package endpoint provides utilities for setting up and handling API endpoints.
package endpoint

import (
	"easyflow-oauth2-server/internal/errors"
	"easyflow-oauth2-server/internal/tokens"
	e "errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Error definitions.
var (
	ErrUserNotFoundError = e.New("user not found in context")
	ErrUserTypeError     = e.New("type assertion to *tokens.JWTTokenPayload failed")
)

type options struct {
	withBody bool
	withUser bool
}

// Option defines a function type for configuring endpoint options.
type Option func(opts *options)

// Utils holds the extracted utilities for an endpoint.
// T is a generic type parameter representing the expected payload type.
// If no payload is expected, use 'any' as the type argument.
type Utils[T any] struct {
	Payload T
	User    *tokens.JWTTokenPayload
}

func getPayload[T any](c *gin.Context) (T, error) {
	var payload T

	if err := c.ShouldBind(&payload); err != nil {
		return payload, err
	}

	return payload, nil
}

func getUser(c *gin.Context) (*tokens.JWTTokenPayload, error) {
	rawUser, ok := c.Get("user")
	if !ok {
		return nil, ErrUserNotFoundError
	}

	user, ok := rawUser.(*tokens.JWTTokenPayload)
	if !ok {
		return nil, ErrUserTypeError
	}

	return user, nil
}

// WithoutBody is an option to indicate that the endpoint does not expect a request body.
func WithoutBody() Option {
	return func(opts *options) {
		opts.withBody = false
	}
}

// WithUser is an option to indicate that the endpoint expects a user in the context.
func WithUser() Option {
	return func(opts *options) {
		opts.withUser = true
	}
}

// SetupEndpoint is a helper function to extract and validate common components from the Gin context.
// It returns an EndpointUtils struct containing the extracted components and a slice of error messages, if any.
// The function takes a generic type parameter T for the expected payload type, provide any if no payload is expected.
func SetupEndpoint[T any](c *gin.Context, opts ...Option) (Utils[T], []string) {
	// Default options
	options := &options{
		withBody: true,
		withUser: false,
	}

	// Apply provided options
	for _, opt := range opts {
		opt(options)
	}

	var errs []error
	var endpointUtils Utils[T]

	// Extract payload
	if options.withBody {
		payload, err := getPayload[T](c)
		if err != nil {
			errs = append(errs, err)
		}
		endpointUtils.Payload = payload
	}

	// Extract User (if available)
	if options.withUser {
		user, err := getUser(c)
		if err != nil {
			errs = append(errs, err)
		}
		endpointUtils.User = user
	}

	var serializableErrors []string

	for _, e := range errs {
		if validationError, ok := e.(validator.ValidationErrors); ok {
			errArr := errors.TranslateError(validationError)
			serializableErrors = append(serializableErrors, errArr...)
		} else {
			serializableErrors = append(serializableErrors, e.Error())
		}
	}

	return endpointUtils, serializableErrors
}

// SendSetupErrorResponse is a helper function to send a standardized error response.
func SendSetupErrorResponse(c *gin.Context, errs []string) {
	errors.SendErrorResponse(c, http.StatusInternalServerError, errors.InternalServerError, errs)
}
