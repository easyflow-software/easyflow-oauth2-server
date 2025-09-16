package endpoint

import (
	"context"
	"crypto/ed25519"
	"easyflow-oauth2-server/pkg/api/errors"
	"easyflow-oauth2-server/pkg/config"
	"easyflow-oauth2-server/pkg/database"
	"easyflow-oauth2-server/pkg/logger"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/valkey-io/valkey-go"
)

type EndpointUtils[T any] struct {
	Payload        T
	Logger         *logger.Logger
	Queries        *database.Queries
	RequestContext context.Context
	Config         *config.Config
	Valkey         valkey.Client
	Key            *ed25519.PrivateKey
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
	raw_queries, ok := c.Get("queries")
	if !ok {
		return nil, fmt.Errorf("queries not found in context")
	}

	queries, ok := raw_queries.(*database.Queries)
	if !ok {
		return nil, fmt.Errorf("type assertion to *database.Queries failed")
	}

	return queries, nil
}

func getConfig(c *gin.Context) (*config.Config, error) {
	raw_cfg, ok := c.Get("config")
	if !ok {
		return nil, fmt.Errorf("config not found in context")
	}

	cfg, ok := raw_cfg.(*config.Config)
	if !ok {
		return nil, fmt.Errorf("type assertion to *common.Config failed")
	}

	return cfg, nil
}

func getLogger(c *gin.Context) (*logger.Logger, error) {
	raw_logger, ok := c.Get("logger")
	if !ok {
		return nil, fmt.Errorf("logger not found in context")
	}

	logger, ok := raw_logger.(*logger.Logger)
	if !ok {
		return nil, fmt.Errorf("type assertion to *logger.Logger failed")
	}

	return logger, nil
}

func getValkey(c *gin.Context) (valkey.Client, error) {
	raw_valkey, ok := c.Get("valkey")
	if !ok {
		return nil, fmt.Errorf("valkey not found in context")
	}

	valkey, ok := raw_valkey.(valkey.Client)
	if !ok {
		return nil, fmt.Errorf("type assertion to *valkey.Client failed")
	}

	return valkey, nil
}

func getKey(c *gin.Context) (*ed25519.PrivateKey, error) {
	raw_key, ok := c.Get("key")
	if !ok {
		return nil, fmt.Errorf("key not found in context")
	}

	key, ok := raw_key.(*ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("type assertion to *ed25519.PrivateKey failed")
	}

	return key, nil
}

func SetupEndpoint[T any](c *gin.Context) (EndpointUtils[T], []string) {
	var errs []error
	var endpointUtils EndpointUtils[T]
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
