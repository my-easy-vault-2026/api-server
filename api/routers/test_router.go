package routers

import (
	"api-server/api/handlers/test"
	middleware "api-server/api/middlewares"
	"api-server/lib"

	"github.com/gin-gonic/gin"
)

type TestRouter struct {
	r                      *gin.Engine `name:"api"`
	apiAuthorityMiddleWare *middleware.ApiAuthorityMiddleWare
	env                    *lib.Env
	logger                 lib.Logger
	accountHandler         *test.AccountHandler
	testHandler            *test.TestHandler
}

var _ IRouter = (*TestRouter)(nil)

func NewTestRouter(r *gin.Engine, apiAuthorityMiddleWare *middleware.ApiAuthorityMiddleWare, logger lib.Logger) *TestRouter {
	return &TestRouter{
		r:                      r,
		apiAuthorityMiddleWare: apiAuthorityMiddleWare,
		logger:                 logger,
	}
}

func (tr *TestRouter) Setup() {

	if tr.env.Environment == "local" ||
		tr.env.Environment == "dev" {
		t := tr.r.Group("/test")
		{

			t.POST("/account/addAssets", tr.accountHandler.AddAssets)
			t.POST("/test/forTest", tr.testHandler.ForTest)

		}
	}
}
