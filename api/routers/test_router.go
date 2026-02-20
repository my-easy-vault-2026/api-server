package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/my-easy-vault-2026/api-server/api/handlers/test"
	middleware "github.com/my-easy-vault-2026/api-server/api/middlewares"
	"github.com/my-easy-vault-2026/api-server/docs"
	"github.com/my-easy-vault-2026/api-server/infra"
	"github.com/my-easy-vault-2026/api-server/lib"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
		tr.apiRouter.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		tr.apiRouter.GET("/docs/doc.json", func(ctx *gin.Context) {
			ctx.Writer.Header().Set("Content-Type", "application/json")
			ctx.Writer.WriteHeader(200)
			ctx.Writer.Write([]byte(docs.SwaggerInfo.ReadDoc()))
		})
		t := tr.apiRouter.ApiGroup.Group("/test")
		{

			t.POST("/account/addAssets", tr.accountHandler.AddAssets)
			t.POST("/test/forTest", tr.testHandler.ForTest)

		}
	}
}
