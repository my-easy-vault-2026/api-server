package test

import (
	"api-server/services"

	"github.com/gin-gonic/gin"
)

type TestHandler struct {
	testService *services.TestService
}

func NewTestHandler() *TestHandler {
	return &TestHandler{
		testService: services.NewTestService(),
	}
}

// @Router			/test/test/forTest [post]
// @Tags			test/test
func (ch *TestHandler) ForTest(c *gin.Context) {

	err := ch.testService.ForTest(c)
	c.JSON(200, err)
}
