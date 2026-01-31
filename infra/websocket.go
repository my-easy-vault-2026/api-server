package infra

import (
	"api-server/lib"
	"api-server/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"shared-modules/common"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type WsServer struct {
	ServerID        string
	Buckets         []*Bucket
	handlers        map[common.MsgOPCode]func(context.Context, *common.Msg) error
	WriteWait       time.Duration
	PongWait        time.Duration
	PingPeriod      time.Duration
	broadcast       chan *common.Msg
	MaxMessageSize  int64
	ReadBufferSize  int
	WriteBufferSize int
	BroadcastSize   int64
	logger          lib.Logger
	env             *lib.Env
	redis           Redis
}

type Bucket struct {
	cLock     sync.RWMutex        // protect the channels for chs
	channels  map[uint64]*Channel // bucket room channels
	broadcast chan *common.Msg
}

type Channel struct {
	broadcast  chan *common.Msg
	userID     uint64
	Conn       *websocket.Conn
	ConnTcp    *net.TCPConn
	CancelFunc context.CancelFunc
}

func NewChannel(size int64) (c *Channel) {
	c = new(Channel)
	c.broadcast = make(chan *common.Msg, size)
	return
}

func (ch *Channel) Push(ctx context.Context, msg *common.Msg) (err error) {
	select {
	case ch.broadcast <- msg:
	default:
	}
	return
}

func NewBucket(size int64) (b *Bucket) {
	b = new(Bucket)
	b.channels = make(map[uint64]*Channel)
	b.broadcast = make(chan *common.Msg, size)
	return
}

func (b *Bucket) Channel(userID uint64) (ch *Channel) {
	b.cLock.RLock()
	ch = b.channels[userID]
	b.cLock.RUnlock()
	return
}

func (b *Bucket) Put(userID uint64, ch *Channel) (err error) {
	b.cLock.Lock()
	ch.userID = userID
	b.channels[userID] = ch
	b.cLock.Unlock()
	return
}

func NewWs(rds Redis, env *lib.Env, logger lib.Logger) *WsServer {

	bucketSize := env.BucketSize
	writeWait := env.WriteWait * time.Millisecond
	pongWait := env.PongWait * time.Millisecond
	pingPeriod := env.PingPeriod * time.Millisecond
	maxMessageSize := env.MaxMessageSize
	readBufferSize := env.ReadBufferSize
	writeBufferSize := env.WriteBufferSize
	broadcastSize := env.BroadcastSize

	ws := &WsServer{
		ServerID:        env.GoNode,
		Buckets:         make([]*Bucket, bucketSize),
		handlers:        make(map[common.MsgOPCode]func(context.Context, *common.Msg) error),
		WriteWait:       writeWait,
		PongWait:        pongWait,
		PingPeriod:      pingPeriod,
		broadcast:       make(chan *common.Msg, broadcastSize),
		MaxMessageSize:  maxMessageSize,
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
		BroadcastSize:   broadcastSize,
		logger:          logger,
		env:             env,
		redis:           rds,
	}

	for i := range ws.Buckets {
		ws.Buckets[i] = NewBucket(ws.BroadcastSize)
	}

	return ws
}

// reduce lock competition, use google city hash insert to different bucket
func (s *WsServer) Bucket(userID uint64) *Bucket {
	userIDStr := fmt.Sprintf("%d", userID)
	idx := utils.CityHash32([]byte(userIDStr), uint32(len(userIDStr))) % uint32(len(s.Buckets))
	return s.Buckets[idx]
}

func (s *WsServer) WritePump(ctx context.Context, ch *Channel) {
	//PingPeriod default eq 54s
	ticker := time.NewTicker(s.PingPeriod)
	defer func() {
		ticker.Stop()
		ch.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-ch.broadcast:
			//write data dead time , like http timeout , default 10s
			ch.Conn.SetWriteDeadline(time.Now().Add(s.WriteWait))
			if !ok {
				s.logger.Warn("SetWriteDeadline not ok")
				ch.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := ch.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				s.logger.Warn(" ch.Conn.NextWriter err :%s  ", err.Error())
				return
			}
			message.NodeID = s.ServerID
			message.MsgID = utils.Md5String(time.Now().String())
			if message.SequenceID == "" {
				message.SequenceID = message.MsgID
			}
			s.logger.Infof("message write body:[%d][%s][%s]", message.OP, message.MsgID, message.Msg)
			j, err := json.Marshal(message)
			if err != nil {
				s.logger.Error("json marshal failed, ", err)
				continue
			}
			w.Write(j)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			//heartbeat，if ping error will exit and close current websocket Conn
			ch.Conn.SetWriteDeadline(time.Now().Add(s.WriteWait))
			s.logger.Debugf("websocket.PingMessage :%v", websocket.PingMessage)
			if err := ch.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

			err := s.redis.Expire(ctx, utils.GetWebsocketNodeKey(ch.userID), s.PingPeriod*2).Err()
			if errors.Is(err, redis.Nil) {
				err = s.redis.Set(ctx, utils.GetWebsocketNodeKey(ch.userID), s.env.GoNode, s.PingPeriod*2).Err()
				if err != nil {
					s.logger.Error("set failed: ", err)
					return
				}
			} else if err != nil {
				s.logger.Warn("expire err : %s", err.Error())
				return
			}
		}
	}
}

func (s *WsServer) ReadPump(ctx context.Context, ch *Channel) {

	for {
		_, message, err := ch.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.Errorf("readPump ReadMessage err:%s", err.Error())
				return
			}
		}
		if message == nil {
			return
		}
		msg := &common.Msg{}
		s.logger.Infof("get a message :%s", message)
		if err := json.Unmarshal([]byte(message), msg); err != nil {
			s.logger.Errorf("unmarshal failed [%s], ", message, err)
			return
		}
		s.handle(ctx, msg)
	}
}

func (s *WsServer) RegisterHandler(opCode common.MsgOPCode, handler func(context.Context, *common.Msg) error) {
	s.handlers[opCode] = handler
}

func (s *WsServer) handle(ctx context.Context, msg *common.Msg) {

	if h, ok := s.handlers[msg.OP]; ok {
		err := h(ctx, msg)
		if err != nil {
			s.logger.Warnf("handler err: %v ", err)
		}
	} else {
		s.logger.Infof("no such OP code[%d][%#v]", msg.OP, msg)
	}

}
