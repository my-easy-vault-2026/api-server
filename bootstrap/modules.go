package bootstrap

import (
	"api-server/api/handlers"
	"api-server/api/middlewares"
	"api-server/api/routers"
	"api-server/dao"
	"api-server/infra"
	"api-server/lib"
	"api-server/services"

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
)
