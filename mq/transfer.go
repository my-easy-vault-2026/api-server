package mq

import (
	"context"

	"github.com/my-easy-vault-2026/shared-modules/common"

	"github.com/my-easy-vault-2026/api-server/services"
)

type TransferHandler struct {
	websocketService *services.WebsocketService
}

func NewTransferHandler(websocketService *services.WebsocketService) *TransferHandler {
	return &TransferHandler{
		websocketService: websocketService,
	}
}

func (ch *TransferHandler) ForwardMessage(ctx context.Context, msg *common.Msg) error {

	if err := ch.websocketService.ForwardMessage(ctx, msg); err != nil {
		return err
	}

	return nil
}
