package web

import (
	"api-server/infra"
	"api-server/lib"
	"api-server/services"
	"context"
	"errors"
	"net/http"
	"shared-modules/common"
	"shared-modules/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type WebsocketHandler struct {
	websocketService *services.WebsocketService
	logger           lib.Logger
	redis            infra.Redis
	wsServer         infra.WsServer
	beBuilder        *lib.BEBuilder
	httpRes          *lib.HttpRes
}

func NewWebsocketHandler(websockerService *services.WebsocketService,
	logger lib.Logger,
	redis infra.Redis,
	wsServer infra.WsServer,
	beBuilder *lib.BEBuilder,
	httpRes *lib.HttpRes,
) *WebsocketHandler {
	return &WebsocketHandler{
		websocketService: websockerService,
		logger:           logger,
		redis:            redis,
		wsServer:         wsServer,
		beBuilder:        beBuilder,
		httpRes:          httpRes,
	}
}

// @Tags			web/websocket
// @Param			X-Token			header	string						true	"User token"
// @Param			Accept-Language	header	string						false	"accept language"
// @Router			/web/websocket/connect/*token [get]
func (wh *WebsocketHandler) Connect(c *gin.Context) {

	token := c.Param("token")
	token = strings.ReplaceAll(token, "/", "")
	wsKey := utils.GetWsTokenRedisKey(common.ROLE_USER, token)

	ret := wh.redis.Get(c, wsKey)
	if errors.Is(redis.Nil, ret.Err()) {
		wh.httpRes.ReError(c, http.StatusBadRequest, wh.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}
	if ret.Err() != nil {
		wh.logger.Warn("get failed", ret.Err())
		wh.httpRes.ReError(c, http.StatusBadRequest, wh.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	userIDStr := ret.Val()
	if userIDStr == "" {
		wh.logger.Error("no X-Uid")
		wh.httpRes.ReError(c, http.StatusBadRequest, wh.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		wh.logger.Error("X-Uid parse failed,", err)
		wh.httpRes.ReError(c, http.StatusBadRequest, wh.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	var upGrader = websocket.Upgrader{
		ReadBufferSize:  int(wh.wsServer.ReadBufferSize),
		WriteBufferSize: int(wh.wsServer.WriteBufferSize),
	}
	//cross origin domain support
	upGrader.CheckOrigin = func(r *http.Request) bool { return true }

	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		wh.logger.Errorf("serverWs err:%s", err.Error())
		conn.Close()
		return
	}

	if err = wh.websocketService.Connect(context.Background(), conn, userID); err != nil {
		conn.Close()
	}
}
