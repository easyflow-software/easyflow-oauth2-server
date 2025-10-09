// Package service provides base service functionalities and dependencies.
package service

import (
	"database/sql"
	"easyflow-oauth2-server/pkg/config"
	"easyflow-oauth2-server/pkg/database"
	"easyflow-oauth2-server/pkg/logger"

	"github.com/valkey-io/valkey-go"
	"go.uber.org/fx"
)

// Option defines a functional option for configuring services.
type Option[T any] func(*T)

// BaseService provides common dependencies and functionality for all services.
type BaseService struct {
	Name          string
	Config        *config.Config
	LoggerFactory *logger.Factory
	DB            *sql.DB
	Queries       *database.Queries
	Valkey        valkey.Client
}

// BaseServiceParams holds the dependencies for BaseService.
type BaseServiceParams struct {
	fx.In
	Config        *config.Config
	LoggerFactory *logger.Factory
	DB            *sql.DB
	Queries       *database.Queries
	Valkey        valkey.Client
}

// NewBaseService creates a new instance of BaseService with the provided parameters.
func NewBaseService(name string, params BaseServiceParams) *BaseService {
	return &BaseService{
		Name:          name,
		Config:        params.Config,
		LoggerFactory: params.LoggerFactory,
		DB:            params.DB,
		Queries:       params.Queries,
		Valkey:        params.Valkey,
	}
}

// GetLogger returns a logger instance for the service with the specified IP.
func (s *BaseService) GetLogger(ip string) *logger.Logger {
	return s.LoggerFactory.NewLogger(ip)
}
