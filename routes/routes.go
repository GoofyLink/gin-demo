package routes

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"reddit/logger"
)

func Signup() *gin.Engine {
	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinRecovery(true))
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	return r
}
