package tasks

import (
	"api-server/services"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"

	"github.com/gin-gonic/gin"
)

// 定时任务由gocron 服务通过定时调用http接口来实现 定时任务url前缀为/tasks/xxx
type CallbackJob struct {
	callbackService *services.NotifyService
}

func NewCallbackJob() *CallbackJob {
	return &CallbackJob{
		callbackService: services.NewNotifyService(),
	}
}

// @Summary		RetryCallback
// @Description	RetryCallback
// @Param			request	body	entities.RetryCallbackForm	true	"body"
// @Tags			tasks
// @Router			/tasks/callback/retry [post]
func (cj *CallbackJob) RetryCallback(c *gin.Context) {

	form := &entities.RetryCallbackForm{}

	err := c.ShouldBindJSON(form)

	if err != nil {
		utils.ReError(c, err)
		return
	}
	if len(form.Types) == 0 {
		form.Types = []common.CallbackType{
			common.CALLBACK_TYPE_MERCHANT_PAY,
			common.CALLBACK_TYPE_MERCHANT_3DS,
			common.CALLBACK_TYPE_MERCHANT_WALLET_OTP,
			common.CALLBACK_TYPE_MERCHANT_CARD_STATUS,
		}
	}

	for _, t := range form.Types {
		err = cj.callbackService.RetryCallback(c, t)
		if err != nil {
			logger.Warnf("callback failed type:[%d], ", t, err)
			continue
		}
	}

	utils.ReData(
		c,
		nil,
	)
}
