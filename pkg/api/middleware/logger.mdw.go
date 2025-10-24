package middleware

import (
	"easyflow-oauth2-server/pkg/api/errors"
	"easyflow-oauth2-server/pkg/config"
	"easyflow-oauth2-server/pkg/logger"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// Creates a new logger instance and adds it to the Gin context.
// It requires the config middleware to be run first to access logging configuration.
// If config is not found or invalid, it aborts the request with a 500 error.
func LoggerMiddleware(module_name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg, ok := c.Get("config")
		if !ok {
			c.JSON(http.StatusInternalServerError, errors.ApiError{
				Code:    http.StatusInternalServerError,
				Error:   "ConfigError",
				Details: "Config not found in context",
			})
			c.Abort()
			return
		}

		config, ok := cfg.(*config.Config)
		if !ok {
			c.JSON(http.StatusInternalServerError, errors.ApiError{
				Code:    http.StatusInternalServerError,
				Error:   "ConfigError",
				Details: "Config is not of type *common.Config",
			})
			c.Abort()
			return
		}

		c.Set("logger", logger.NewLogger(os.Stdout, module_name, config.LogLevel, c.ClientIP()))
		c.Next()
	}
}
