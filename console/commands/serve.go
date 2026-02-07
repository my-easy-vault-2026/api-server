package commands

import (
	"api-server/api/middlewares"
	"api-server/api/routers"
	"api-server/infra"
	"api-server/lib"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"
)

// ServeCommand test command
type ServeCommand struct{}

func (s *ServeCommand) Short() string {
	return "serve application"
}

func (s *ServeCommand) Setup(cmd *cobra.Command) {}

func (s *ServeCommand) Run() lib.CommandRunner {
	return func(
		middleware middlewares.Middlewares,
		env *lib.Env,
		router infra.Router,
		route routers.Routers,
		logger lib.Logger,
		database infra.Database,

	) {
		logger.Info(`+-----------------------+`)
		logger.Info(`| EASY VAULT API SERVER |`)
		logger.Info(`+-----------------------+`)

		// Using time zone as specified in env file
		loc, _ := time.LoadLocation(env.TimeZone)
		time.Local = loc

		middleware.Setup()
		route.Setup()

		if env.Environment != "local" && env.SentryDSN != "" {
			err := sentry.Init(sentry.ClientOptions{
				Dsn:              env.SentryDSN,
				AttachStacktrace: true,
			})
			if err != nil {
				logger.Error("sentry initialization failed")
				logger.Error(err.Error())
			}
		}
		logger.Info("Running server")
		if env.ServerPort == "" {
			if err := router.Run(); err != nil {
				logger.Fatal(err)
				return
			}
		} else {
			if err := router.Run(":" + env.ServerPort); err != nil {
				logger.Fatal(err)
				return
			}
		}
	}
}

func NewServeCommand() *ServeCommand {
	return &ServeCommand{}
}
