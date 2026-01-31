package websocket

import (
	"api-server/lib"
	"api-server/services"
	"context"
	"shared-modules/common"
)

type CommonHandler struct {
	websocketService *services.WebsocketService
	logger           lib.Logger
}

func NewCommonHandler(websocketService *services.WebsocketService, logger lib.Logger) *CommonHandler {
	return &CommonHandler{
		websocketService: websocketService,
		logger:           logger,
	}
}

func (ch *CommonHandler) Read(ctx context.Context, msg *common.Msg) error {
	if err := ch.websocketService.Read(ctx, msg); err != nil {
		return err
	}
	return nil
}
