package infra

import (
	"api-server/lib"
	"context"
	"errors"
	"sync"
	"time"
)

type Locker struct {
	key      string
	locked   bool
	rwLock   sync.RWMutex
	lockChan chan struct{}
	errChan  chan error
	logger   lib.Logger
	redis    Redis
}

type Lockers struct {
	logger lib.Logger
	redis  Redis
}

func NewLockers(logger lib.Logger, redis Redis) Lockers {
	return Lockers{logger: logger, redis: redis}
}

func (l *Lockers) NewLocker() *Locker {
	return &Locker{
		lockChan: make(chan struct{}, 0),
		errChan:  make(chan error, 0),
		logger:   l.logger,
		redis:    l.redis,
	}
}

func (l *Locker) WaitLock(ctx context.Context, key string, duration time.Duration, wait time.Duration) error {

	l.key = key
	l.rwLock.Lock()
	defer l.rwLock.Unlock()

	cctx, cancel := context.WithTimeout(
		ctx,
		time.Duration(wait),
	)

	go func(cctx context.Context) {
		for {
			if !l.locked {
				ret := l.redis.SetNX(cctx, key, time.Now().Unix(), duration)
				if err := ret.Err(); err != nil {
					select {
					case l.errChan <- err:
						cancel()
						return
					case <-time.After(100 * time.Microsecond):
						l.logger.Warn("setnx err chan block")
						cancel()
						return
					}
				}
				if ret.Val() {
					select {
					case l.lockChan <- struct{}{}:
						cancel()
						return
					case <-time.After(100 * time.Microsecond):
						l.logger.Warn("setnx res chan block")
						cancel()
						return
					}
				}
			}
			time.Sleep(100 * time.Microsecond)
		}
	}(cctx)
	defer cancel()

	select {
	case err := <-l.errChan:
		if err != nil {
			return err
		}
	case <-l.lockChan:
		return nil
	case <-cctx.Done():
		switch cctx.Err() {
		case context.DeadlineExceeded, context.Canceled:
			return errors.New("system busy")
		}
		return errors.New("system busy")
	case <-time.After(wait):
		l.logger.Warnf("wait lock timeout [%s]", key)
		return errors.New("system busy")
	}

	return errors.New("unknown error")
}

func (l *Locker) Lock(ctx context.Context, key string, duration time.Duration) bool {

	l.key = key
	l.rwLock.Lock()
	defer l.rwLock.Unlock()

	ret := l.redis.SetNX(ctx, key, time.Now().Unix(), duration)

	if ret.Err() != nil {
		return false
	}

	return ret.Val()

}

func (l *Locker) UnLock(ctx context.Context) error {

	l.rwLock.Lock()
	defer l.rwLock.Unlock()

	ret := l.redis.Del(ctx, l.key)

	if ret.Err() != nil {
		return ret.Err()
	}
	return nil
}
