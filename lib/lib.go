package lib

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewEnv),
	fx.Provide(GetLogger),
	fx.Provide(NewI18N),
	fx.Provide(NewHttpRes),
)
