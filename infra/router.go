package infra

import (
	"net/http"

	"github.com/my-easy-vault-2026/api-server/lib"
	"github.com/my-easy-vault-2026/api-server/utils"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Router -> Gin Router
type Router struct {
	*gin.Engine
}

// NewRouter : all the routes are defined here
func NewAPIRouter(
	env *lib.Env,
	logger lib.Logger,
) Router {

	if env.Environment != "local" && env.SentryDSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:         env.SentryDSN,
			Environment: `api-server-` + env.Environment,
		}); err != nil {
			logger.Infof("Sentry initialization failed: %v\n", err)
		}
	}

	gin.DefaultWriter = logger.GetGinLogger()
	appEnv := env.Environment
	if appEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	httpRouter := gin.Default()

	httpRouter.MaxMultipartMemory = env.MaxMultipartMemory

	httpRouter.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "OPTIONS", "DELETE"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))

	// Attach sentry middleware
	httpRouter.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
	}))

	httpRouter.GET("/health-check", func(c *gin.Context) {
		utils.SendSentryMsg(c, "Error")
		c.JSON(http.StatusOK, gin.H{"data": "API Up and Running"})
	})

	return Router{
		httpRouter,
	}
}

func NewWebsocketRouter(
	env *lib.Env,
	logger lib.Logger,
) Router {

	if env.Environment != "local" && env.SentryDSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:         env.SentryDSN,
			Environment: `api-server-` + env.Environment,
		}); err != nil {
			logger.Infof("Sentry initialization failed: %v\n", err)
		}
	}

	gin.DefaultWriter = logger.GetGinLogger()
	appEnv := env.Environment
	if appEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	httpRouter := gin.Default()

	httpRouter.MaxMultipartMemory = env.MaxMultipartMemory

	httpRouter.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "OPTIONS", "DELETE"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))

	// Attach sentry middleware
	httpRouter.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
	}))

	httpRouter.GET("/health-check", func(c *gin.Context) {
		utils.SendSentryMsg(c, "Error")
		c.JSON(http.StatusOK, gin.H{"data": "API Up and Running"})
	})

	return Router{
		httpRouter,
	}
}
