// Package container provides fx modules for dependency injection
package container

import (
	"easyflow-oauth2-server/internal/routes/admin"
	"easyflow-oauth2-server/internal/routes/auth"
	"easyflow-oauth2-server/internal/routes/oauth"
	"easyflow-oauth2-server/internal/routes/user"

	"go.uber.org/fx"
)

// ServicesModule contains all the service providers.
var ServicesModule = fx.Module("services",
	fx.Provide(
		// Auth services
		auth.NewAuthService,
		auth.NewAuthController,

		// OAuth services
		oauth.NewOAuthService,
		oauth.NewOAuthController,

		// Admin services
		admin.NewAdminService,
		admin.NewAdminController,

		// User services
		user.NewUserService,
		user.NewUserController,
	),
)
