package infra

import (
	"context"
	"strconv"
	"sync"

	"shared-modules/common"
	"shared-modules/logger"
)

var (
	workerPoolTasks map[common.WorkerPool]chan func() error
	workerPoolLock  sync.RWMutex
	workerPoolOnce  sync.Once
)

func RegisterWorkerPool(pool common.WorkerPool, size int) {
	workerPoolOnce.Do(func() {
		workerPoolTasks = make(map[common.WorkerPool]chan func() error)
	})

	ok := func() bool {
		workerPoolLock.RLock()
		defer workerPoolLock.RUnlock()
		_, ok := workerPoolTasks[pool]
		return ok
	}()

	if !ok {
		workerPoolLock.Lock()
		defer workerPoolLock.Unlock()
		_, ok := workerPoolTasks[pool]
		if !ok {
			tasks := make(chan func() error, Config.System.WorkerPoolSize)
			workerPoolTasks[pool] = tasks
			for i := 0; i < size; i++ {
				go func() {
					SetNonExpiringMDCValue("reqId", "wp-"+pool.String()+strconv.Itoa(i))
					for task := range tasks {
						err := task()
						if err != nil {
							logger.Warnf("worker pool task err: [%v]", err)
						}
					}
				}()
			}
		}
	}
}

func SubmitWorkerPool(ctx context.Context, pool common.WorkerPool, task func() error) error {
	workerPoolOnce.Do(func() {
		workerPoolTasks = make(map[common.WorkerPool]chan func() error)
	})

	wp, ok := func() (chan func() error, bool) {
		workerPoolLock.RLock()
		defer workerPoolLock.RUnlock()
		wp, ok := workerPoolTasks[pool]
		return wp, ok
	}()
	if !ok {
		return NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	wp <- task
	return nil
}
