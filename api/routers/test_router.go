package main

import (
	"api-server/api/handlers/test"
	"api-server/lib"

	"github.com/gin-gonic/gin"
)

type TestRouter struct {
	r              *gin.Engine `name:"api"`
	env            *lib.Env
	logger         lib.Logger
	accountHandler *test.AccountHandler
	testHandler    *test.TestHandler
}

func NewTestRouter(r *gin.Engine, apiAuthorityMiddleWare *middleware.ApiAuthorityMiddleWare, logger lib.Logger) *TestRouter {
	return &TestRouter{
		r:                      r,
		apiAuthorityMiddleWare: apiAuthorityMiddleWare,
		logger:                 logger,
	}
}

func (tr *TestRouter) SetRoute() {

	if tr.env.Environment == "local" ||
		tr.env.Environment == "dev" {
		t := tr.r.Group("/test")
		{

			t.POST("/account/addAssets", tr.accountHandler.AddAssets)
			t.POST("/test/forTest", tr.testHandler.ForTest)

		}
	}
}
