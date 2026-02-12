package middlewares

import (
	"net/http"
	"runtime/debug"

	"github.com/my-easy-vault-2026/shared-modules/common"

	"github.com/my-easy-vault-2026/api-server/infra"
	"github.com/my-easy-vault-2026/api-server/lib"
	"github.com/my-easy-vault-2026/shared-modules/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type RecoverMiddleWare struct {
	logger          lib.Logger
	beBuilder       *lib.BEBuilder
	httpRes         *lib.HttpRes
	APIRouter       infra.Router `name:"api"`
	WebsocketRouter infra.Router `name:"websocket"`
}

type RecoverMiddleWareParams struct {
	fx.In
	Logger          lib.Logger
	BEBuilder       *lib.BEBuilder
	HTTPRes         *lib.HttpRes
	APIRouter       infra.Router `name:"api"`
	WebsocketRouter infra.Router `name:"websocket"`
}

func NewRecoverMiddleWare(
	p RecoverMiddleWareParams,
) *RecoverMiddleWare {
	return &RecoverMiddleWare{
		logger:          p.Logger,
		beBuilder:       p.BEBuilder,
		httpRes:         p.HTTPRes,
		APIRouter:       p.APIRouter,
		WebsocketRouter: p.WebsocketRouter,
	}
}

func (rm *RecoverMiddleWare) Setup() {
	rm.APIRouter.Use(rm.Handle())
	rm.WebsocketRouter.Use(rm.Handle())
}

func (rm *RecoverMiddleWare) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {

				traceID, ok := c.Get(common.HEADER_X_TRACE_ID)
				if ok {
					c.Writer.Header().Set(common.HEADER_X_TRACE_ID, traceID.(string))
				}

				path := c.Request.URL.RequestURI()
				file, line, fn := utils.FormatStackOneLineWithCode()
				rm.logger.Errorf("panic recovered on [%s][%s],%v", path, fn, err)

				rm.logger.Errorf("SEARCH_CODE:%s|%d", file, line)
				rm.logger.Warnf(string(debug.Stack()))
				rm.httpRes.ReError(c,
					http.StatusInternalServerError,
					rm.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
			}
		}()
		c.Next()
	}
}
