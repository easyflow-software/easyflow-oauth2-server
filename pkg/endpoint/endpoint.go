// Package endpoint provides utilities for setting up and handling API endpoints.
package endpoint

import (
	"context"
	"crypto/ed25519"
	"easyflow-oauth2-server/pkg/api/errors"
	"easyflow-oauth2-server/pkg/config"
	"easyflow-oauth2-server/pkg/database"
	"easyflow-oauth2-server/pkg/logger"
	"easyflow-oauth2-server/pkg/tokens"
	e "errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/valkey-io/valkey-go"
)

// Error definitions.
var (
	ErrQueryNotFoundError  = e.New("queries not found in context")
	ErrQueryTypeError      = e.New("type assertion to *database.Queries failed")
	ErrConfigNotFoundError = e.New("config not found in context")
	ErrConfigTypeError     = e.New("type assertion to *config.Config failed")
	ErrLoggerNotFoundError = e.New("logger not found in context")
	ErrLoggerTypeError     = e.New("type assertion to *logger.Logger failed")
	ErrValkeyNotFoundError = e.New("valkey not found in context")
	ErrValkeyTypeError     = e.New("type assertion to valkey.Client failed")
	ErrKeyNotFoundError    = e.New("key not found in context")
	ErrKeyTypeError        = e.New("type assertion to *ed25519.PrivateKey failed")
	ErrUserNotFoundError   = e.New("user not found in context")
	ErrUserTypeError       = e.New("type assertion to *tokens.JWTTokenPayload failed")
)

// Utils holds the extracted utilities for an endpoint.
// T is a generic type parameter representing the expected payload type.
// If no payload is expected, use 'any' as the type argument.
type Utils[T any] struct {
	Payload        T
	Logger         *logger.Logger
	Queries        *database.Queries
	RequestContext context.Context
	Config         *config.Config
	Valkey         valkey.Client
	Key            *ed25519.PrivateKey
	User           *tokens.JWTTokenPayload
}

func getPayload[T any](c *gin.Context) (T, error) {
	var payload T

	if err := c.ShouldBind(&payload); err != nil && err.Error() != "EOF" {
		// Handle binding errors, but ignore io.EOF which occurs when the body is empty
		return payload, err
	}

	return payload, nil
}

func getQueries(c *gin.Context) (*database.Queries, error) {
	rawQueries, ok := c.Get("queries")
	if !ok {
		return nil, ErrQueryNotFoundError
	}

	queries, ok := rawQueries.(*database.Queries)
	if !ok {
		return nil, ErrQueryTypeError
	}

	return queries, nil
}

func getConfig(c *gin.Context) (*config.Config, error) {
	rawCfg, ok := c.Get("config")
	if !ok {
		return nil, ErrConfigNotFoundError
	}

	cfg, ok := rawCfg.(*config.Config)
	if !ok {
		return nil, ErrConfigTypeError
	}

	return cfg, nil
}

func getLogger(c *gin.Context) (*logger.Logger, error) {
	rawLogger, ok := c.Get("logger")
	if !ok {
		return nil, ErrLoggerNotFoundError
	}

	logger, ok := rawLogger.(*logger.Logger)
	if !ok {
		return nil, ErrLoggerTypeError
	}

	return logger, nil
}

func getValkey(c *gin.Context) (valkey.Client, error) {
	rawValkey, ok := c.Get("valkey")
	if !ok {
		return nil, ErrValkeyNotFoundError
	}

	valkey, ok := rawValkey.(valkey.Client)
	if !ok {
		return nil, ErrValkeyTypeError
	}

	return valkey, nil
}

func getKey(c *gin.Context) (*ed25519.PrivateKey, error) {
	rawKey, ok := c.Get("key")
	if !ok {
		return nil, ErrKeyNotFoundError
	}

	key, ok := rawKey.(*ed25519.PrivateKey)
	if !ok {
		return nil, ErrKeyTypeError
	}

	return key, nil
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

// SetupEndpoint is a helper function to extract and validate common components from the Gin context.
// It returns an EndpointUtils struct containing the extracted components and a slice of error messages, if any.
// The function takes a generic type parameter T for the expected payload type, provide any if no payload is expected.
func SetupEndpoint[T any](c *gin.Context) (Utils[T], []string) {
	var errs []error
	var endpointUtils Utils[T]
	// Extract Request Context
	endpointUtils.RequestContext = c.Request.Context()

	// Extract payload
	payload, err := getPayload[T](c)
	if err != nil {
		errs = append(errs, err)
	}
	endpointUtils.Payload = payload

	// Extract Queries
	queries, err := getQueries(c)
	if err != nil {
		errs = append(errs, err)
	}
	endpointUtils.Queries = queries

	// Extract Config
	cfg, err := getConfig(c)
	if err != nil {
		errs = append(errs, err)
	}
	endpointUtils.Config = cfg

	// Extract Logger
	logger, err := getLogger(c)
	if err != nil {
		errs = append(errs, err)
	}
	endpointUtils.Logger = logger

	// Extract Valkey Client
	valkeyClient, err := getValkey(c)
	if err != nil {
		errs = append(errs, err)
	}
	endpointUtils.Valkey = valkeyClient

	// Extract Ed25519 Key
	key, err := getKey(c)
	if err != nil {
		errs = append(errs, err)
	}
	endpointUtils.Key = key

	// Extract User (if available)
	user, err := getUser(c)
	if err != nil {
		// User might not be set for public endpoints, so we don't append the error
		endpointUtils.User = nil
	} else {
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
