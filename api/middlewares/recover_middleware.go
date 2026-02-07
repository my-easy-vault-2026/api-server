package middlewares

import (
	"api-server/lib"
	"api-server/utils"
	"net/http"
	"runtime/debug"
	"shared-modules/common"

	"github.com/gin-gonic/gin"
)

type RecoverMiddleWare struct {
	logger    lib.Logger
	beBuilder *lib.BEBuilder
	httpRes   *lib.HttpRes
}

func NewRecoverMiddleWare(logger lib.Logger,
	beBuilder *lib.BEBuilder,
	httpRes *lib.HttpRes) *RecoverMiddleWare {
	return &RecoverMiddleWare{
		logger:    logger,
		beBuilder: beBuilder,
		httpRes:   httpRes,
	}
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
