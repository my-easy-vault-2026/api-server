package mq

import (
	"api-server/infra"
	"shared-modules/common"

	"go.uber.org/fx"
)

type MQs struct {
	mq              *infra.MQ
	transferHandler *TransferHandler
}

type IMQ interface {
	Setup()
}

var Module = fx.Options(
	fx.Provide(NewTransferHandler),
	fx.Provide(NewMQs),
)

func NewMQs(mq *infra.MQ,
	transferHandler *TransferHandler) *MQs {
	return &MQs{
		mq:              mq,
		transferHandler: transferHandler,
	}
}

func (m *MQs) Setup() {
	m.mq.RegisterHandler(common.MSG_OPCODE_INFUND, m.transferHandler.ForwardMessage)
}
