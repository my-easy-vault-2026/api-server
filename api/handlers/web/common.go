package web

import (
	"net/http"

	"github.com/my-easy-vault-2026/shared-modules/common"

	"github.com/my-easy-vault-2026/api-server/lib"

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
