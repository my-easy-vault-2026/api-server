package mq

import (
	"context"
	"fmt"

	"api-server/services"
	"shared-modules/common"
	"shared-modules/utils"
)

type CardHandler struct {
	websocketService *services.WebsocketService
}

func NewCardHandler() *CardHandler {
	return &CardHandler{
		websocketService: services.NewWebsocketService(),
	}
}

func (ch *CardHandler) ForwardThreedsToInstance(ctx context.Context, msg *common.Msg) error {

	userIDStr := fmt.Sprintf("%d", msg.UserID)
	idx := utils.CityHash32([]byte(userIDStr), uint32(len(userIDStr))) % uint32(utils.Config.System.NodeSize)
	msg.OP = common.MSG_OPCODE_FORWARD_3DS_BALANCED
	if err := utils.MQUtil.Push(ctx, utils.GetPubsubKey("websocket", fmt.Sprintf("%d", idx)), msg); err != nil {
		return err
	}
	return nil
}

func (ch *CardHandler) ForwardThreeds(ctx context.Context, msg *common.Msg) error {

	if err := ch.websocketService.ForwardThreeds(ctx, msg); err != nil {
		return err
	}

	return nil
}
