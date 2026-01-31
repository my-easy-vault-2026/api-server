package main

import (
	"api-server/api/handlers/web"
	"api-server/lib"
	"context"
	"shared-modules/common"
	"shared-modules/utils"

	"github.com/gin-gonic/gin"
	"github.com/polevpn/elog"
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
	websocketHandler       *web.WebSocketHandler
}

func NewWebRouter(r *gin.Engine, apiAuthorityMiddleWare *middleware.ApiAuthorityMiddleWare, logger lib.Logger) *WebRouter {
	return &WebRouter{
		r:                      r,
		apiAuthorityMiddleWare: apiAuthorityMiddleWare,
		logger:                 logger,
	}
}

func (wr *WebRouter) SetRoute() {

	w := wr.r.Group("/web", wr.apiAuthorityMiddleWare.Handle())
	{
		w.POST("/user/loginOrRegister", wr.authHandler.LoginOrRegister)

		w.POST("/user/logout", wr.authHandler.Logout)
		w.POST("/user/getInfo", wr.userHandler.GetInfo)
		w.GET("/wallet/list", utils.VRHandle(wr.commonHandler.VersionOutdated,
			utils.NewVRFunc("1.0.0", "1.1.9", wr.commonHandler.VersionOutdated),
			utils.NewVRFunc("1.1.9", "9.9.9", wr.walletHandler.ListWallets),
		))
		w.GET("/wallet/category", utils.VRHandle(wr.commonHandler.VersionOutdated,
			utils.NewVRFunc("1.0.0", "1.1.9", wr.commonHandler.VersionOutdated),
			utils.NewVRFunc("1.1.9", "9.9.9", wr.walletHandler.ListCategory),
		))
		w.POST("/wallet/apply", wr.walletHandler.CreateWallet)

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

	if utils.Config.Server.Env == "local" ||
		utils.Config.Server.Env == "ngrok" ||
		utils.Config.Server.Env == "dev" ||
		utils.Config.Server.Env == "test" {
		t := r.Group("/test")
		{
			testHandler := test.NewTestHandler()
			accountHandler := test.NewAccountHandler()

			t.POST("/account/addAssets", accountHandler.AddAssets)
			t.POST("/test/forTest", testHandler.ForTest)

		}
	}
}

func SubscribeMQ() {

	if err := utils.InitMQ(context.Background(),
		[]string{
			utils.GetMsgListKey("api"),
			utils.GetMsgListKey("api", utils.EnvConfig.GoNode),
			utils.GetQueueKey("open", "log"),
			utils.GetQueueKey("open", "log", utils.EnvConfig.GoNode),
			utils.GetQueueKey("merchant", "account", "export_balance_change"),
		},
		[]string{utils.GetPubsubKey("api")},
	); err != nil {
		elog.Fatal("init mq failed,", err)
	}
}

func SubscribeWS() {
	commonHandler := websocket.NewCommonHandler()
	utils.Ws.RegisterHandler(common.MSG_OPCODE_READ, commonHandler.Read)
}
