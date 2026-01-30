package main

import (
	"api-server/handlers/test"
	"api-server/handlers/web"
	"api-server/handlers/websocket"
	"context"
	"shared-modules/common"
	"shared-modules/utils"

	"github.com/gin-gonic/gin"
	"github.com/polevpn/elog"
)

func SetRoute(r *gin.Engine, rw *gin.Engine) {

	// CORS 设置
	// r.Use(cors.New(cors.Config{
	// 	AllowOrigins:     []string{"http://localhost"},
	// 	AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	// 	AllowHeaders:     []string{"Origin", "X-Requested-With", "Content-Type", "Accept", "Authorization", "X-Token"},
	// 	AllowCredentials: true,
	// 	ExposeHeaders:    []string{"Set-Cookie"},
	// }))

	w := r.Group("/web", CheckAuth)
	{
		accountHandler := web.NewAccountHandler()
		authHandler := web.NewAuthHandler()
		userHandler := web.NewUserHandler()
		cardHandler := web.NewCardHandler()
		quoteHandler := web.NewQuoteHandler()
		walletHandler := web.NewWalletHandler()
		orderHandler := web.NewOrderHandler()
		// exchangeHandler := web.NewExchangeHandler()
		transferHandler := web.NewTransferHandler()
		commonHandler := web.NewCommonHandler()
		financialHandler := web.NewFinancialHandler()

		w.POST("/account/assets/getEquivalent", accountHandler.GetEquivalentAsset)

		w.POST("/user/loginOrRegister", authHandler.LoginOrRegister)

		w.POST("/user/logout", authHandler.Logout)
		w.POST("/user/setPinCode", userHandler.SetPinCode)
		w.POST("/user/resetPinCode", utils.VRHandle(commonHandler.VersionOutdated, utils.NewVRFunc("1.1.10", "9.9.9", userHandler.ResetPinCode)))
		w.POST("/user/forgotPinCode", utils.VRHandle(commonHandler.VersionOutdated, utils.NewVRFunc("1.1.10", "9.9.9", userHandler.ForgotPinCode)))
		w.POST("/user/getInfo", userHandler.GetInfo)
		w.POST("/user/savePhoneNumber", userHandler.SavePhoneNumber)
		w.POST("/user/deleteAccount", userHandler.DeleteAccount)

		w.POST("/card/list", utils.VRHandle(cardHandler.ListCard,
			utils.NewVRFunc("1.0.0", "1.1.9", cardHandler.ListCryptpAndProductCard),
			utils.NewVRFunc("1.1.9", "9.9.9", cardHandler.ListCard),
		))
		w.POST("/card/get", cardHandler.GetCard)
		w.POST("/card/listCategory", utils.VRHandle(cardHandler.ListCardCategory,
			utils.NewVRFunc("1.3.9", "9.9.9", cardHandler.ListCardCategory),
		))

		// this api will be deprecated
		w.POST("/quote/exchangeRates/list", quoteHandler.ListExchangeRate)

		w.POST("/quote/exchangeRates/getRates", quoteHandler.GetExchange)

		w.POST("/wallet/list", walletHandler.ListWallets)
		w.POST("/wallet/apply", walletHandler.CreateWallet)

		w.POST("/order/transactions/page", orderHandler.PageTransactionRecords)
		w.POST("/order/transactions/get", orderHandler.GetTransactionRecord)
		w.POST("/order/transactions/autoYield/page", orderHandler.PageAutoYieldTransactionRecords)

		// w.POST("/exchange/preview", exchangeHandler.ExchangePreview) // 暫不開放
		// w.POST("/exchange/confirm", exchangeHandler.ExchangeConfirm) // 暫不開放

		w.POST("/transfer/preview", transferHandler.TransferPreview)
		w.POST("/transfer/confirm", transferHandler.TransferConfirm)

		w.POST("/financial/autoYield/info", financialHandler.AutoYieldInfo)
		w.POST("/financial/autoYield/history", financialHandler.AutoYieldHistory)
		w.POST("/financial/autoYield/enable", financialHandler.AutoYieldEnable)
		w.POST("/financial/autoYield/list", financialHandler.AutoYieldList)
	}

	ww := rw.Group("/web", WsCheckAuth)
	{
		websocketHandler := web.NewWebsocketHandler()
		ww.GET("/websocket/connect/*token", websocketHandler.Connect)
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
