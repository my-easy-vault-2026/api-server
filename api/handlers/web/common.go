package web

import (
	"api-server/lib"
	"net/http"
	"shared-modules/common"

	"github.com/gin-gonic/gin"
)

type CommonHandler struct {
	logger    lib.Logger
	beBuilder *lib.BEBuilder
	httpRes   *lib.HttpRes
}

func NewCommonHandler(logger lib.Logger, beBuilder *lib.BEBuilder, httpRes *lib.HttpRes) *CommonHandler {
	return &CommonHandler{
		logger:    logger,
		beBuilder: beBuilder,
		httpRes:   httpRes,
	}
}

func (ch *CommonHandler) VersionOutdated(c *gin.Context) {
	ch.httpRes.ReError(c, http.StatusUpgradeRequired, ch.beBuilder.NewBusinessError(c, common.CODE_APP_VERSION_OUTDATED))
	return
}
