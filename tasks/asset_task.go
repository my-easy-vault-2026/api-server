package tasks

import (
	"api-server/lib"
	"api-server/services"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// 定时任务由gocron 服务通过定时调用http接口来实现 定时任务url前缀为/tasks/xxx
type AssetTaskService struct {
	accountService *services.AccountService
	logger         lib.Logger
}

func NewAssetTaskService(
	accountService *services.AccountService,
	logger lib.Logger,
) *AssetTaskService {
	return &AssetTaskService{
		accountService: accountService,
		logger:         logger,
	}
}

// @Summary 手動轉帳
// @Description 手動轉帳
// @Tags tasks
// @Param request body entities.ManualTransferForm true "body"
// @Router /tasks/asset/manualTransfer [post]
func (tj *AssetTaskService) ManualTransfer(c *gin.Context) {

	form := &entities.ManualTransferForm{}

	err := c.ShouldBindJSON(form)

	if err != nil {
		utils.ReError(c, err)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(form); err != nil {
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_REQUEST_BODY_INVALID_FORMAT, err.Error()))
		return
	}

	orderNO, err := tj.accountService.ManualTransfer(c, form)

	if err != nil {
		utils.ReError(c, err)
	}

	logger.Infof("manual transfer success [%s]", orderNO)

	utils.ReData(
		c,
		orderNO,
	)
}

// @Tags			tasks
// @Router			/tasks/asset/snapshot [post]
func (tj *AssetTaskService) AssetSnapshot(c *gin.Context) {

	err := tj.accountService.DailyAssetSnapshot(c)
	if err != nil {
		logger.Warn("daily asset snapshot failed,", err)
		return
	}

}
