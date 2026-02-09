package routers

import (
	"api-server/api/handlers/test"
	middleware "api-server/api/middlewares"
	"api-server/infra"
	"api-server/lib"

	"go.uber.org/fx"
)

type TestRouter struct {
	apiRouter              infra.Router `name:"api"`
	apiAuthorityMiddleWare *middleware.ApiAuthorityMiddleWare
	env                    *lib.Env
	logger                 lib.Logger
	accountHandler         *test.AccountHandler
	testHandler            *test.TestHandler
}

type TestRouterParams struct {
	fx.In
	ApiRouter              infra.Router `name:"api"`
	ApiAuthorityMiddleWare *middleware.ApiAuthorityMiddleWare
	Env                    *lib.Env
	Logger                 lib.Logger
	AccountHandler         *test.AccountHandler
	TestHandler            *test.TestHandler
}

var _ IRouter = (*TestRouter)(nil)

func NewTestRouter(
	p TestRouterParams,
) *TestRouter {
	return &TestRouter{
		apiRouter:              p.ApiRouter,
		apiAuthorityMiddleWare: p.ApiAuthorityMiddleWare,
		logger:                 p.Logger,
		accountHandler:         p.AccountHandler,
		testHandler:            p.TestHandler,
		env:                    p.Env,
	}
}

func (tr *TestRouter) Setup() {

	if tr.env.Environment == "local" ||
		tr.env.Environment == "dev" {
		t := tr.apiRouter.Group("/test")
		{

			t.POST("/account/addAssets", tr.accountHandler.AddAssets)
			t.POST("/test/forTest", tr.testHandler.ForTest)

		}
	}
}
