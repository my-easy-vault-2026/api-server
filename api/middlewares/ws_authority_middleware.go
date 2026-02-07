package middlewares

import (
	"api-server/lib"
	"api-server/services"

	"github.com/gin-gonic/gin"
)

type WebsocketAuthorityMiddleWare struct {
	authService *services.AuthService
	logger      lib.Logger
}

func NewWebsocketAuthorityMiddleWare(authService *services.AuthService, logger lib.Logger) *WebsocketAuthorityMiddleWare {
	return &WebsocketAuthorityMiddleWare{
		authService: authService,
		logger:      logger,
	}
}

func (ah *WebsocketAuthorityMiddleWare) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {

		// TODO: websocket auth logic

		c.Next()
	}
}
