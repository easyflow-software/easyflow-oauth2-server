// Package container provides fx modules for dependency injection
package container

import (
	"go.uber.org/fx"
)

// AppModule contains the complete application setup.
var AppModule = fx.Module("app",
	ProvidersModule,
	ServicesModule,
	ServerModule,
)

// NewApp creates a new fx application with all modules.
func NewApp() *fx.App {
	return fx.New(
		AppModule,
		fx.NopLogger, // Disable fx's own logging to avoid noise
	)
}
