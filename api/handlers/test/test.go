package test

import (
	"github.com/my-easy-vault-2026/api-server/lib"
	"github.com/my-easy-vault-2026/api-server/services"

	"github.com/gin-gonic/gin"
)

type TestHandler struct {
	testService *services.TestService
	logger      lib.Logger
}

func NewTestHandler(testService *services.TestService, logger lib.Logger) *TestHandler {
	return &TestHandler{
		testService: testService,
		logger:      logger,
	}
}

// @Router			/test/test/forTest [post]
// @Tags			test/test
func (ch *TestHandler) ForTest(c *gin.Context) {

	err := ch.testService.ForTest(c)
	c.JSON(200, err)
}
