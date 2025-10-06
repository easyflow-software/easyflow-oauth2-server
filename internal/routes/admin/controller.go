// Package admin contains admin-related endpoints
package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Controller handles admin HTTP requests.
type Controller struct {
	service *Service
}

// NewAdminController creates a new instance of AdminController.
func NewAdminController(service *Service) *Controller {
	return &Controller{
		service: service,
	}
}

// RegisterRoutes sets up the admin-related endpoints.
func (ctrl *Controller) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/system-info", ctrl.GetSystemInfo)
	r.GET("/stats", ctrl.GetStats)
}

// GetSystemInfo handles requests for system information.
func (ctrl *Controller) GetSystemInfo(c *gin.Context) {
	systemInfo, err := ctrl.service.GetSystemInfo(c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve system info"})
		return
	}

	c.JSON(http.StatusOK, systemInfo)
}

// GetStats handles requests for system statistics.
func (ctrl *Controller) GetStats(c *gin.Context) {
	stats, err := ctrl.service.GetStats(c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
