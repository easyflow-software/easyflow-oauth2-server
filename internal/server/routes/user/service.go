package user

import (
	"easyflow-oauth2-server/internal/service"

	"go.uber.org/fx"
)

// Service handles user-related business logic.
type Service struct {
	*service.BaseService
}

// ServiceParams holds dependencies for UserService.
type ServiceParams struct {
	fx.In
	service.BaseServiceParams
}

// NewUserService creates a new instance of UserService.
func NewUserService(params ServiceParams) *Service {
	baseService := service.NewBaseService("UserService", params.BaseServiceParams)
	return &Service{
		BaseService: baseService,
	}
}
