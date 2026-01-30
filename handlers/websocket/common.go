package websocket

import (
	"api-server/services"
	"context"
	"shared-modules/common"
)

type CommonHandler struct {
	websocketService *services.WebsocketService
}

func NewCommonHandler() *CommonHandler {
	return &CommonHandler{
		websocketService: services.NewWebsocketService(),
	}
}

func (ch *CommonHandler) Read(ctx context.Context, msg *common.Msg) error {
	if err := ch.websocketService.Read(ctx, msg); err != nil {
		return err
	}
	return nil
}
