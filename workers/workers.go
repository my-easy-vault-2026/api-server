package workers

import (
	"api-server/infra"
	"api-server/lib"
	"shared-modules/common"

	"go.uber.org/fx"
)

type Workers struct {
	env          *lib.Env
	workderPools *infra.WorkerPools
}

type IWorkers interface {
	Setup()
}

var Module = fx.Options(
	fx.Provide(NewWorkers),
)

func NewWorkers(env *lib.Env, workderPools *infra.WorkerPools) *Workers {
	return &Workers{
		env:          env,
		workderPools: workderPools,
	}
}

func (w *Workers) Setup() {
	w.workderPools.RegisterWorkerPool(common.WORKER_POOL_DEFAULT, w.env.DefaultWorkerSize)
}
