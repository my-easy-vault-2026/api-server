package infra

import (
	"api-server/lib"
	"api-server/utils"
	"context"
	"encoding/json"
	"errors"
	"runtime/debug"

	"github.com/redis/go-redis/v9"

	"shared-modules/common"

	"shared-modules/logger"
)

type MQ struct {
	redis    Redis
	handlers map[common.MsgOPCode]func(context.Context, *common.Msg) error
	logger   lib.Logger
	env      *lib.Env
}

func NewMQ(ctx context.Context, rds Redis, logger lib.Logger, env *lib.Env) (*MQ, error) {
	mq := &MQ{
		redis:    rds,
		handlers: make(map[common.MsgOPCode]func(context.Context, *common.Msg) error),
		logger:   logger,
		env:      env,
	}

	var err error
	var lists, pubsubs []string
	lists = env.MqList
	pubsubs = env.MqPubsub

	for _, l := range lists {
		go func() {
			ctx := context.WithValue(ctx, "reqId", "mq-"+l)
			for {
				var result []string
				result, err = mq.redis.BRPop(ctx, 1000, l).Result()
				if err != nil && !errors.Is(err, redis.Nil) {
					logger.Debugf("brpop failed:%v", err)
				}
				if err != nil {
					continue
				}

				if len(result) >= 2 {
					mq.handle(ctx, result[1])
				}
			}
		}()
	}

	for _, p := range pubsubs {
		go func() {
			ctx := context.WithValue(ctx, "reqId", "mq-"+p)
			subscriber := mq.redis.Subscribe(ctx, p)
			for {
				msg, err := subscriber.ReceiveMessage(ctx)
				if err != nil {
					logger.Debugf("receive failed:%v", err)
					continue
				}

				mq.handle(ctx, msg.Payload)
			}
		}()
	}
	return mq, nil
}

func (m *MQ) RegisterHandler(opCode common.MsgOPCode, handler func(context.Context, *common.Msg) error) {
	m.handlers[opCode] = handler
}

func (m *MQ) Push(ctx context.Context, queueName string, msg *common.Msg) error {
	j, err := json.Marshal(msg)
	if err != nil {
		logger.Warnf("unmarshal err:[%#v] [%v] ", msg, err)
		return err
	}
	logger.Infof("push msg info [%d][%#v]", msg.OP, msg)

	if err := m.redis.LPush(ctx, queueName, j).Err(); err != nil {
		logger.Warnf("lpush err:%v", err)
		return err
	}
	return nil
}

func (m *MQ) Pub(ctx context.Context, queueName string, msg *common.Msg) error {
	j, err := json.Marshal(msg)
	if err != nil {
		logger.Warnf("unmarshal err:[%#v] [%v] ", msg, err)
		return err
	}
	logger.Infof("publish msg info [%d][%#v]", msg.OP, msg)

	if err := m.redis.Publish(ctx, queueName, j).Err(); err != nil {
		logger.Warnf("publish err:%v", err)
		return err
	}
	return nil
}

func (m *MQ) handle(ctx context.Context, msgstr string) {
	var opCode string

	defer func() {
		if err := recover(); err != nil {

			file, line, fn := utils.FormatStackOneLineWithCode()
			logger.Errorf("panic recovered on [%s][%s],%v", opCode, fn, err)

			logger.Errorf("SEARCH_CODE:%s|%d", file, line)
			logger.Warnf(string(debug.Stack()))
		}
	}()
	msg := &common.Msg{}
	if err := json.Unmarshal([]byte(msgstr), msg); err != nil {
		logger.Infof("unmarshal err:[]%s %v ", msgstr, err)
	}
	opCode = string(msg.OP)
	logger.Infof("push msg info [%d][%#v]", msg.OP, msg)

	if h, ok := m.handlers[msg.OP]; ok {
		err := h(ctx, msg)
		if err != nil {
			logger.Warnf("handler err: %v ", err)
		}
	} else {
		logger.Warnf("no such OP code[%d][%#v]", msg.OP, msg)
	}
}
