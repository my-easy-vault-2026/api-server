package bootstrap

import (
	"github.com/my-easy-vault-2026/api-server/api/handlers"
	"github.com/my-easy-vault-2026/api-server/api/middlewares"
	"github.com/my-easy-vault-2026/api-server/api/routers"
	"github.com/my-easy-vault-2026/api-server/dao"
	"github.com/my-easy-vault-2026/api-server/infra"
	"github.com/my-easy-vault-2026/api-server/lib"
	"github.com/my-easy-vault-2026/api-server/mq"
	"github.com/my-easy-vault-2026/api-server/services"
	"github.com/my-easy-vault-2026/api-server/workers"

	"go.uber.org/fx"
)

var CommonModules = fx.Options(
	handlers.Module,
	routers.Module,
	services.Module,
	dao.Module,
	infra.Module,
	middlewares.Module,
	lib.Module,
	mq.Module,
	workers.Module,
)
