// Package user contains user-related endpoints and controllers.
package user

import (
	"github.com/gin-gonic/gin"
)

// Controller handles user HTTP requests.
type Controller struct {
	service *Service
}

// NewUserController creates a new instance of UserController.
func NewUserController(service *Service) *Controller {
	return &Controller{
		service: service,
	}
}

// RegisterRoutes sets up the user-related endpoints.
func (ctrl *Controller) RegisterRoutes(_ *gin.RouterGroup) {

}
