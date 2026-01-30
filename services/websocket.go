package services

import (
	"api-server/lib"
	"context"
	"shared-modules/common"
	"shared-modules/logger"
	"shared-modules/utils"
	"time"

	"github.com/gorilla/websocket"
)

type WebsocketService struct {
	logger lib.Logger
}

func NewWebsocketService(logger lib.Logger) *WebsocketService {

	return &WebsocketService{
		logger: logger,
	}
}

func (ws *WebsocketService) Connect(ctx context.Context, conn *websocket.Conn, userID uint64) error {

	var ch *utils.Channel
	//default broadcast size eq 512
	ch = utils.NewChannel(utils.Ws.BroadcastSize)
	ch.Conn = conn

	ctx, cancel := context.WithCancel(context.Background())
	ch.CancelFunc = cancel

	//send data to websocket conn
	go utils.Ws.WritePump(ctx, ch)
	//get data from websocket conn
	go utils.Ws.ReadPump(ctx, ch)

	err := utils.Ws.Bucket(userID).Put(userID, ch)
	if err != nil {
		return utils.NewBusinessError(ctx, common.CODE_WEBSOCKET_CREATE_WEBSOCKET_CHANNEL_FAILED)
	}

	err = utils.RDB.Set(ctx, utils.GetWebsocketNodeKey(userID), utils.EnvConfig.GoNode, utils.Config.Websocket.PingPeriodMs*time.Millisecond*2).Err()
	if err != nil {
		return err
	}

	for {

	}
}

func (ws *WebsocketService) ForwardMessage(ctx context.Context, msg *common.Msg) error {

	c := utils.Ws.Bucket(msg.UserID).Channel(msg.UserID)
	if c == nil {
		logger.Debugf("channel not found: [%d] %#v", msg.UserID, msg)
		return nil
	}

	if err := c.Push(ctx, msg); err != nil {
		return err
	}
	return nil
}

func (ws *WebsocketService) Read(ctx context.Context, msg *common.Msg) error {

	logger.Infof("message read: [%s][%s]", msg.SequenceID, msg.MsgID)
	return nil
}
