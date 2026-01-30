package web

import (
	"api-server/lib"
	"api-server/services"
	"context"
	"errors"
	"net/http"
	"shared-modules/common"
	"shared-modules/logger"
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
}

func NewWebsocketHandler(websockerService *services.WebsocketService, logger lib.Logger) *WebsocketHandler {
	return &WebsocketHandler{
		websocketService: websockerService,
		logger:           logger,
	}
}

// @Summary		Get all wallets
// @Description	Get all wallets of the user
// @Tags			web/websocket
// @Param			X-Token			header	string						true	"User token"
// @Param			Accept-Language	header	string						false	"accept language"
// @Router			/web/websocket/connect [get]
func (wh *WebsocketHandler) Connect(c *gin.Context) {

	token := c.Param("token")
	token = strings.ReplaceAll(token, "/", "")
	wsKey := utils.GetWsTokenRedisKey(common.ROLE_USER, token)

	ret := utils.RDB.Get(c, wsKey)
	if errors.Is(redis.Nil, ret.Err()) {
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_NOT_LOGIN))
		return
	}
	if ret.Err() != nil {
		logger.Warn("get failed", ret.Err())
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	userIDStr := ret.Val()
	if userIDStr == "" {
		logger.Error("no X-Uid")
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		logger.Error("X-Uid parse failed,", err)
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	var upGrader = websocket.Upgrader{
		ReadBufferSize:  int(utils.Ws.ReadBufferSize),
		WriteBufferSize: int(utils.Ws.WriteBufferSize),
	}
	//cross origin domain support
	upGrader.CheckOrigin = func(r *http.Request) bool { return true }

	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		logger.Errorf("serverWs err:%s", err.Error())
		conn.Close()
		return
	}

	if err = wh.websocketService.Connect(context.Background(), conn, userID); err != nil {
		conn.Close()
	}
}
