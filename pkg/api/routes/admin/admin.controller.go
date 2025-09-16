package admin

import (
	"easyflow-oauth2-server/pkg/api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAdminEndpoints(r *gin.RouterGroup) {
	r.Use(middleware.LoggerMiddleware("Admin"))
	
}