package web

import (
	"shared-modules/common"
	"shared-modules/utils"

	"github.com/gin-gonic/gin"
)

type CommonHandler struct {
}

func NewCommonHandler() *CommonHandler {
	return &CommonHandler{}
}

func (ch *CommonHandler) VersionOutdated(c *gin.Context) {
	utils.ReError(c, utils.NewBusinessError(c, common.CODE_APP_VERSION_OUTDATED))
	return
}
