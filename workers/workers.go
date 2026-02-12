package workers

import (
	"github.com/my-easy-vault-2026/shared-modules/common"

	"github.com/my-easy-vault-2026/api-server/lib"

	"github.com/my-easy-vault-2026/api-server/infra"

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
