package mq

import (
	"context"

	"api-server/services"
	"shared-modules/common"
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
