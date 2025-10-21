// Package container provides fx modules for dependency injection
package container

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"easyflow-oauth2-server/internal/server/config"
	"easyflow-oauth2-server/internal/server/routes/admin"
	"easyflow-oauth2-server/internal/server/routes/auth"
	"easyflow-oauth2-server/internal/server/routes/oauth"
	"easyflow-oauth2-server/internal/server/routes/user"
	"easyflow-oauth2-server/internal/server/routes/wellknown"
	"easyflow-oauth2-server/pkg/logger"

	cors "github.com/OnlyNico43/gin-cors/v2"
	"github.com/gin-gonic/gin"
	"github.com/valkey-io/valkey-go"
	"go.uber.org/fx"
)

// ServerModule contains the HTTP server setup.
var ServerModule = fx.Module("server",
	fx.Provide(NewGinRouter),
	fx.Invoke(RegisterRoutes),
)

// RouterParams holds the dependencies for the Gin router.
type RouterParams struct {
	fx.In
	Config        *config.Config
	LoggerFactory *logger.Factory
}

// RouteParams holds all controllers for route registration.
type RouteParams struct {
	fx.In
	Config              *config.Config
	LoggerFactory       *logger.Factory
	Router              *gin.Engine
	AuthController      *auth.Controller
	OAuthController     *oauth.Controller
	AdminController     *admin.Controller
	UserController      *user.Controller
	WellKnownController *wellknown.Controller
	DB                  *sql.DB
	ValkeyClient        valkey.Client
}

// NewGinRouter creates and configures a new Gin router.
func NewGinRouter(params RouterParams) *gin.Engine {
	// Disable Gin debug logs
	gin.SetMode(gin.ReleaseMode)

	// Initialize Gin router
	router := gin.New()

	// Configure trusted proxies
	err := router.SetTrustedProxies(params.Config.TrustedProxies)
	if err != nil {
		log := params.LoggerFactory.NewLogger("System")
		log.PrintfError("Could not set trusted proxies list: %v", err)
	}

	// Configure router path handling
	router.RedirectFixedPath = true     // Redirect to the correct path if case-insensitive match found
	router.RedirectTrailingSlash = true // Automatically handle trailing slashes

	// Add recovery middleware
	router.Use(gin.Recovery())

	return router
}

// RegisterRoutes registers all application routes.
func RegisterRoutes(lc fx.Lifecycle, params RouteParams) {
	log := params.LoggerFactory.NewLogger("System")

	// Set up CORS middleware
	corsMiddleware := cors.Middleware(&cors.Config{
		AllowedOriginsFunc: func(origin string) (bool, bool) {
			if origin == params.Config.FrontendURL {
				return true, true
			}
			return true, false
		},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:  []string{"Content-Length"},
		MaxAge:         12 * time.Hour,
	})

	params.Router.Use(corsMiddleware)

	// Register admin routes
	adminEndpoints := params.Router.Group("/admin")
	log.PrintfInfo("Registering admin endpoints")
	params.AdminController.RegisterRoutes(adminEndpoints)

	// Register auth routes
	authEndpoints := params.Router.Group("/auth")
	log.PrintfInfo("Registering auth endpoints")
	params.AuthController.RegisterRoutes(authEndpoints)

	// Register OAuth routes
	oauthEndpoints := params.Router.Group("/oauth")
	log.PrintfInfo("Registering oauth endpoints")
	params.OAuthController.RegisterRoutes(oauthEndpoints)

	// Register user routes
	userEndpoints := params.Router.Group("/user")
	log.PrintfInfo("Registering user endpoints")
	params.UserController.RegisterRoutes(userEndpoints)

	// Register .well-known routes
	wellKnownEndpoints := params.Router.Group("/.well-known")
	log.PrintfInfo("Registering .well-known endpoints")
	params.WellKnownController.RegisterRoutes(wellKnownEndpoints)

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + params.Config.Port,
		Handler: params.Router,
	}

	// Register lifecycle hooks
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				log.PrintfInfo("Starting Server on port %s", params.Config.Port)
				if err := server.ListenAndServe(); err != nil &&
					!errors.Is(err, http.ErrServerClosed) {
					log.PrintfError("Failed to start server: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.PrintfInfo("Shutting down server...")

			// Close database connection
			if params.DB != nil {
				log.PrintfInfo("Closing database connection")
				if err := params.DB.Close(); err != nil {
					log.PrintfError("Error closing database connection: %v", err)
				}
			}

			// Close Valkey connection
			if params.ValkeyClient != nil {
				log.PrintfInfo("Closing Valkey connection")
				params.ValkeyClient.Close()
			}

			return server.Shutdown(ctx)
		},
	})
}
