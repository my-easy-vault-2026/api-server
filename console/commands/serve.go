package commands

import (
	"api-server/api/middlewares"
	"api-server/api/routers"
	"api-server/infra"
	"api-server/lib"
	"api-server/mq"
	"api-server/workers"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// ServeCommand test command
type ServeCommand struct {
}

type ServeCommandParams struct {
	fx.In
	Middleware      middlewares.Middlewares
	Env             *lib.Env
	ApiRouter       infra.Router `name:"api"`
	WebsocketRouter infra.Router `name:"websocket"`
	Route           routers.Routers
	Logger          lib.Logger
	Database        infra.Database
	Mq              *mq.MQs
	Workers         *workers.Workers
}

func (s *ServeCommand) Short() string {
	return "serve application"
}

func (s *ServeCommand) Setup(cmd *cobra.Command) {}

func (s *ServeCommand) Run() lib.CommandRunner {
	return func(
		p ServeCommandParams,
	) {
		p.Logger.Info(`+-----------------------+`)
		p.Logger.Info(`| EASY VAULT API SERVER |`)
		p.Logger.Info(`+-----------------------+`)

		// Using time zone as specified in env file
		loc, _ := time.LoadLocation(p.Env.TimeZone)
		time.Local = loc

		p.Middleware.Setup()
		p.Route.Setup()
		p.Mq.Setup()
		p.Workers.Setup()

		if p.Env.Environment != "local" && p.Env.SentryDSN != "" {
			err := sentry.Init(sentry.ClientOptions{
				Dsn:              p.Env.SentryDSN,
				AttachStacktrace: true,
			})
			if err != nil {
				p.Logger.Error("sentry initialization failed")
				p.Logger.Error(err.Error())
			}
		}

		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {

			p.Logger.Info("Running server")
			if p.Env.ServerPort == "" {
				if err := p.ApiRouter.Run(); err != nil {
					p.Logger.Fatal(err)
					return
				}
			} else {
				if err := p.ApiRouter.Run(":" + p.Env.ServerPort); err != nil {
					p.Logger.Fatal(err)
					return
				}
			}
			wg.Done()
		}()

		go func() {
			p.Logger.Info("Running webocket server")
			if p.Env.WSServerPort == "" {
				if err := p.WebsocketRouter.Run(); err != nil {
					p.Logger.Fatal(err)
					return
				}
			} else {
				if err := p.WebsocketRouter.Run(":" + p.Env.WSServerPort); err != nil {
					p.Logger.Fatal(err)
					return
				}
			}
			wg.Done()
		}()
		wg.Wait()
	}
}

func NewServeCommand() *ServeCommand {
	return &ServeCommand{}
}
