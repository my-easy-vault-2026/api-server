package services

import (
	"api-server/infra"
	"api-server/lib"
	"context"
	"shared-modules/common"
	"shared-modules/utils"

	"github.com/gorilla/websocket"
)

type WebsocketService struct {
	logger    lib.Logger
	beBuilder *lib.BEBuilder
	ws        *infra.WsServer
	env       *lib.Env
	redis     infra.Redis
}

func NewWebsocketService(logger lib.Logger,
	beBuilder *lib.BEBuilder,
	ws *infra.WsServer,
	env *lib.Env,
	redis infra.Redis) *WebsocketService {

	return &WebsocketService{
		logger:    logger,
		beBuilder: beBuilder,
		ws:        ws,
		env:       env,
		redis:     redis,
	}
}

func (ws *WebsocketService) Connect(ctx context.Context, conn *websocket.Conn, userID uint64) error {

	var ch *infra.Channel
	//default broadcast size eq 512
	ch = infra.NewChannel(ws.env.BroadcastSize)
	ch.Conn = conn

	ctx, cancel := context.WithCancel(context.Background())
	ch.CancelFunc = cancel

	//send data to websocket conn
	go ws.ws.WritePump(ctx, ch)
	//get data from websocket conn
	go ws.ws.ReadPump(ctx, ch)

	err := ws.ws.Bucket(userID).Put(userID, ch)
	if err != nil {
		return ws.beBuilder.NewBusinessError(ctx, common.CODE_CREATE_WEBSOCKET_CHANNEL_FAILED)
	}

	err = ws.redis.Set(ctx, utils.GetWebsocketNodeKey(userID), ws.env.GoNode, ws.env.PingPeriod*2).Err()
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
			// default:
			//     var msg Message
			//     if err := conn.ReadJSON(&msg); err != nil {
			//         return err
			//     }
			//     // 處理消息
			//     ws.handleMessage(ctx, userID, msg)
		}
	}
}

func (ws *WebsocketService) ForwardMessage(ctx context.Context, msg *common.Msg) error {

	c := ws.ws.Bucket(msg.UserID).Channel(msg.UserID)
	if c == nil {
		ws.logger.Debugf("channel not found: [%d] %#v", msg.UserID, msg)
		return nil
	}

	if err := c.Push(ctx, msg); err != nil {
		return err
	}
	return nil
}

func (ws *WebsocketService) Read(ctx context.Context, msg *common.Msg) error {

	ws.logger.Infof("message read: [%s][%s]", msg.SequenceID, msg.MsgID)
	return nil
}
