package test

import (
	"api-server/lib"
	"api-server/services"

	"github.com/gin-gonic/gin"
)

type TestHandler struct {
	testService *services.TestService
	logger      lib.Logger
}

func NewTestHandler(testService *services.TestService, logger lib.Logger) *TestHandler {
	return &TestHandler{
		testService: services.NewTestService(),
		logger:      logger,
	}
}

// @Router			/test/test/forTest [post]
// @Tags			test/test
func (ch *TestHandler) ForTest(c *gin.Context) {

	err := ch.testService.ForTest(c)
	c.JSON(200, err)
}
