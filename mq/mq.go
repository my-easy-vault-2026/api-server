package mq

import (
	"github.com/my-easy-vault-2026/shared-modules/common"

	"github.com/my-easy-vault-2026/api-server/infra"

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
