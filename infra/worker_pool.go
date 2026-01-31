package infra

import (
	"api-server/lib"
	"context"
	"errors"
	"strconv"
	"sync"

	"shared-modules/common"
)

type WorkerPools struct {
	env             *lib.Env
	logger          lib.Logger
	workerPoolTasks map[common.WorkerPool]chan func() error
	workerPoolLock  sync.RWMutex
	workerPoolOnce  sync.Once
}

func NewWorkerPools(env *lib.Env, logger lib.Logger) *WorkerPools {
	wps := &WorkerPools{
		env:             env,
		logger:          logger,
		workerPoolTasks: make(map[common.WorkerPool]chan func() error),
		workerPoolLock:  sync.RWMutex{},
		workerPoolOnce:  sync.Once{},
	}
	return wps
}

func (wps *WorkerPools) RegisterWorkerPool(pool common.WorkerPool, size int) {
	ok := func() bool {
		wps.workerPoolLock.RLock()
		defer wps.workerPoolLock.RUnlock()
		_, ok := wps.workerPoolTasks[pool]
		return ok
	}()

	if !ok {
		wps.workerPoolLock.Lock()
		defer wps.workerPoolLock.Unlock()
		_, ok := wps.workerPoolTasks[pool]
		if !ok {
			tasks := make(chan func() error, wps.env.WorkerPoolSize)
			wps.workerPoolTasks[pool] = tasks
			for i := 0; i < size; i++ {
				go func() {
					for task := range tasks {
						err := task()
						if err != nil {
							wps.logger.Warnf("worker pool task err: [%v]", err)
						}
					}
				}()
			}
		}
	}
}

func (wps *WorkerPools) SubmitWorkerPool(ctx context.Context, pool common.WorkerPool, task func() error) error {
	wps.workerPoolOnce.Do(func() {
		wps.workerPoolTasks = make(map[common.WorkerPool]chan func() error)
	})

	wp, ok := func() (chan func() error, bool) {
		wps.workerPoolLock.RLock()
		defer wps.workerPoolLock.RUnlock()
		wp, ok := wps.workerPoolTasks[pool]
		return wp, ok
	}()
	if !ok {
		return errors.New("worker pool not found: " + strconv.Itoa(int(pool)))
	}
	wp <- task
	return nil
}
