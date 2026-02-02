package router

import (
	"api-server/api/handlers/web"
	middleware "api-server/api/middlewares"
	"api-server/lib"
	"shared-modules/utils"

	"github.com/gin-gonic/gin"
)

type WebRouter struct {
	r                      *gin.Engine `name:"api"`
	rw                     *gin.Engine `name:"websocket"`
	apiAuthorityMiddleWare *middleware.ApiAuthorityMiddleWare
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
}

var _ IRouter = (*WebRouter)(nil)

func NewWebRouter(r *gin.Engine,
	rw *gin.Engine,
	apiAuthorityMiddleWare *middleware.ApiAuthorityMiddleWare,
	logger lib.Logger,
	authHandler *web.AuthHandler,
	userHandler *web.UserHandler,
	quoteHandler *web.QuoteHandler,
	walletHandler *web.WalletHandler,
	orderHandler *web.OrderHandler,
	transferHandler *web.TransferHandler,
	commonHandler *web.CommonHandler,
	exchangeHandler *web.ExchangeHandler,
	websocketHandler *web.WebsocketHandler,
) *WebRouter {
	return &WebRouter{
		r:                      r,
		rw:                     rw,
		apiAuthorityMiddleWare: apiAuthorityMiddleWare,
		logger:                 logger,
		authHandler:            authHandler,
		userHandler:            userHandler,
		quoteHandler:           quoteHandler,
		walletHandler:          walletHandler,
		orderHandler:           orderHandler,
		transferHandler:        transferHandler,
		commonHandler:          commonHandler,
		exchangeHandler:        exchangeHandler,
		websocketHandler:       websocketHandler,
	}
}

func (wr *WebRouter) Setup() {

	w := wr.r.Group("/web", wr.apiAuthorityMiddleWare.Handle())
	{
		w.POST("/auth/loginOrRegister", wr.authHandler.LoginOrRegister)
		w.POST("/auth/logout", wr.authHandler.Logout)

		w.GET("/user/:id", wr.userHandler.GetInfo)

		w.GET("/wallet", utils.VRHandle(wr.commonHandler.VersionOutdated,
			utils.NewVRFunc("1.0.0", "1.1.9", wr.commonHandler.VersionOutdated),
			utils.NewVRFunc("1.1.9", "9.9.9", wr.walletHandler.ListWallets),
		))
		w.GET("/wallet/category", utils.VRHandle(wr.commonHandler.VersionOutdated,
			utils.NewVRFunc("1.0.0", "1.1.9", wr.commonHandler.VersionOutdated),
			utils.NewVRFunc("1.1.9", "9.9.9", wr.walletHandler.ListCategory),
		))
		w.POST("/wallet", wr.walletHandler.CreateWallet)

		// this api will be deprecated
		w.GET("/quote/exchangeRates/list", wr.quoteHandler.ListExchangeRate)

		w.GET("/quote/exchangeRates/getRates", wr.quoteHandler.GetExchange)

		w.POST("/order/transactions/page", wr.orderHandler.PageTransactionRecords)

		w.POST("/exchange/preview", wr.exchangeHandler.ExchangePreview) // 暫不開放
		w.POST("/exchange/confirm", wr.exchangeHandler.ExchangeConfirm) // 暫不開放

		w.POST("/transfer/preview", wr.transferHandler.TransferPreview)
		w.POST("/transfer/confirm", wr.transferHandler.TransferConfirm)
	}

	ww := wr.rw.Group("/web", wr.apiAuthorityMiddleWare.Handle())
	{
		ww.GET("/websocket/connect/*token", wr.websocketHandler.Connect)
	}
}
