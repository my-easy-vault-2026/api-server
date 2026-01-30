package main

import (
	"api-server/docs"
	"flag"
	_ "shared-modules/logger/impl"
	"shared-modules/utils"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/polevpn/elog"
	swaggerfiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
)

var configFile string
var configPath string
var configLangPath string

func init() {
	flag.StringVar(&configFile, "c", "config.yaml", "config file")
	flag.StringVar(&configPath, "p", "conf/", "config path")
	flag.StringVar(&configLangPath, "l", "lang/", "config lang path")
}

func main() {

	var err error
	flag.Parse()
	defer elog.Flush()
	elog.SetMode(elog.WITH_NO_FILE_LINE)

	elog.Info("server start...")

	err = utils.GetConf(configPath + configFile)

	if err != nil {
		elog.Fatal("load config from ", configPath+configFile, " failed,", err)
	}
	gin.SetMode(utils.Config.Gin.Mode)

	err = utils.InitEnv()
	if err != nil {
		elog.Fatal("init env failed,", err)
	}

	for _, db := range utils.Config.Mysql.DBs {
		elog.Debugf("mysql [%s] dsn: %s\n", db.Name, db.Dsn)
		err = utils.RegisterDB(utils.GormMySQLConfig{
			Name:                  db.Name,
			DSN:                   db.Dsn,
			MaxOpenConns:          db.MaxOpenConn,
			MaxIdleConns:          db.MaxIdleConn,
			ConnMaxLifetimeSecond: db.ConMaxLifeTimeSeconds,
		})
		if err != nil {
			elog.Fatal("init db failed,", err)
		}
	}

	err = utils.InitI18N(configLangPath)
	if err != nil {
		elog.Fatal("init i18n failed,", err)
	}

	err = utils.InitRedis(utils.Config.Redis.Addr, utils.Config.Redis.Pwd, utils.Config.Redis.Db, utils.Config.Redis.MaxConn, utils.Config.Redis.TLSVersion)
	if err != nil {
		elog.Fatal("init redis failed,", err)
	}

	utils.InitWs(
		utils.Config.Websocket.BucketSize,
		utils.Config.Websocket.WriteWaitMs*time.Millisecond,
		utils.Config.Websocket.PongWaitMs*time.Millisecond,
		utils.Config.Websocket.PingPeriodMs*time.Millisecond,
		utils.Config.Websocket.MaxMsgSize,
		utils.Config.Websocket.ReadBufferSize,
		utils.Config.Websocket.WriteBufferSize,
		utils.Config.Websocket.BroadcastSize,
	)

	r, rw := gin.Default(), gin.Default()

	r.SetTrustedProxies(utils.Config.Gin.TrustedProxies)
	rw.SetTrustedProxies(utils.Config.Gin.TrustedProxies)

	// 拦截器在这里
	r.Use(Recover)
	r.Use(Logger)
	r.Use(HeaderInjector)
	rw.Use(Recover)
	rw.Use(Logger)
	rw.Use(HeaderInjector)

	SetMiddleware()
	SetRoute(r, rw)
	SubscribeMQ()
	SubscribeWS()

	if utils.Config.Server.Env == "local" || utils.Config.Server.Env == "ngrok" {
		docs.SwaggerInfo.BasePath = "/"
	} else {
		docs.SwaggerInfo.BasePath = "/api"
	}
	docs.SwaggerInfo.Description = "Please insert <b>/api</b> in front of the url."

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		r.Run(utils.Config.Gin.Listen)
	}()
	go func() {
		defer wg.Done()
		rw.Run(utils.Config.Gin.WsListen)
	}()
	wg.Wait()
}
