package routers

import (
	"github.com/my-easy-vault-2026/api-server/api/handlers/web"
	middleware "github.com/my-easy-vault-2026/api-server/api/middlewares"
	"github.com/my-easy-vault-2026/api-server/infra"
	"github.com/my-easy-vault-2026/api-server/lib"

	"go.uber.org/fx"
)

type WebRouter struct {
	apiRouter              infra.Router `name:"api"`
	websocketRouter        infra.Router `name:"websocket"`
	apiAuthorityMiddleWare *middleware.ApiAuthorityMiddleWare
	rateLimitMiddleWare    *middleware.RateLimitMiddleWare
	logger                 lib.Logger
	authHandler            *web.AuthHandler
	userHandler            *web.UserHandler
	quoteHandler           *web.QuoteHandler
	walletHandler          *web.WalletHandler
	orderHandler           *web.OrderHandler
	transferHandler        *web.TransferHandler
	commonHandler          *web.CommonHandler
	exchangeHandler        *web.ExchangeHandler
	websocketHandler       *web.WebsocketHandler
	appVersionLib          *lib.APPVersionLib
}

type WebRouterParams struct {
	fx.In
	ApiRouter              infra.Router `name:"api"`
	WebsocketRouter        infra.Router `name:"websocket"`
	ApiAuthorityMiddleWare *middleware.ApiAuthorityMiddleWare
	RatelimitMiddleware    *middleware.RateLimitMiddleWare
	Logger                 lib.Logger
	AuthHandler            *web.AuthHandler
	UserHandler            *web.UserHandler
	QuoteHandler           *web.QuoteHandler
	WalletHandler          *web.WalletHandler
	OrderHandler           *web.OrderHandler
	TransferHandler        *web.TransferHandler
	CommonHandler          *web.CommonHandler
	ExchangeHandler        *web.ExchangeHandler
	WebsocketHandler       *web.WebsocketHandler
	AppVersionLib          *lib.APPVersionLib
}

var _ IRouter = (*WebRouter)(nil)

func NewWebRouter(
	p WebRouterParams,
) *WebRouter {
	return &WebRouter{
		apiRouter:              p.ApiRouter,
		websocketRouter:        p.WebsocketRouter,
		apiAuthorityMiddleWare: p.ApiAuthorityMiddleWare,
		rateLimitMiddleWare:    p.RatelimitMiddleware,
		logger:                 p.Logger,
		authHandler:            p.AuthHandler,
		userHandler:            p.UserHandler,
		quoteHandler:           p.QuoteHandler,
		walletHandler:          p.WalletHandler,
		orderHandler:           p.OrderHandler,
		transferHandler:        p.TransferHandler,
		commonHandler:          p.CommonHandler,
		exchangeHandler:        p.ExchangeHandler,
		websocketHandler:       p.WebsocketHandler,
		appVersionLib:          p.AppVersionLib,
	}
}

func (wr *WebRouter) Setup() {

	w := wr.apiRouter.Group("/web", wr.apiAuthorityMiddleWare.Handle(), wr.rateLimitMiddleWare.Handle())

	{
		w.POST("/auth/loginOrRegister", wr.authHandler.LoginOrRegister)
		w.POST("/auth/logout", wr.authHandler.Logout)

		w.GET("/user/:id", wr.userHandler.GetInfo)

		w.POST("/wallet", wr.walletHandler.CreateWallet)
		w.GET("/wallet", wr.appVersionLib.VRHandle(wr.commonHandler.VersionOutdated,
			wr.appVersionLib.NewVRFunc("1.0.0", "1.1.8", wr.commonHandler.VersionOutdated),
			wr.appVersionLib.NewVRFunc("1.1.9", "9.9.9", wr.walletHandler.ListWallets),
		))
		w.GET("/wallet/category", wr.appVersionLib.VRHandle(wr.commonHandler.VersionOutdated,
			wr.appVersionLib.NewVRFunc("1.0.0", "1.1.8", wr.commonHandler.VersionOutdated),
			wr.appVersionLib.NewVRFunc("1.1.9", "9.9.9", wr.walletHandler.ListCategory),
		))

		w.GET("/quote/exchangeRate/:quote/:base", wr.quoteHandler.GetExchange)

		w.POST("/order/transaction", wr.orderHandler.PageTransactionRecords)

		w.GET("/exchange/preview", wr.exchangeHandler.ExchangePreview)
		w.POST("/exchange/confirm", wr.exchangeHandler.ExchangeConfirm)

		w.GET("/transfer/preview", wr.transferHandler.TransferPreview)
		w.POST("/transfer/confirm", wr.transferHandler.TransferConfirm)
	}

	ww := wr.websocketRouter.Group("/web")
	{
		ww.GET("/websocket/connect/*token", wr.websocketHandler.Connect)
	}
}
