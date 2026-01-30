package tasks

import (
	"api-server/services"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// 定时任务由gocron 服务通过定时调用http接口来实现 定时任务url前缀为/tasks/xxx
type FinancialTaskJob struct {
	financialService *services.FinancialService
}

func NewFinancialTaskJob() *FinancialTaskJob {
	return &FinancialTaskJob{
		financialService: services.NewFinancialService(),
	}
}

// @Tags			tasks
// @Router			/tasks/financial/autoYield/checkSnapshot [post]
// @Param			request			body		entities.AutoYieldCheckSnapshotForm	false	"body"
func (tj *FinancialTaskJob) AutoYieldCheckSnapshot(c *gin.Context) {

	form := &entities.AutoYieldCheckSnapshotForm{}

	err := c.ShouldBindJSON(&form)
	if err != nil {
		logger.Warnf("unmarshal failed: %v", err)
	}

	now := time.Now().Truncate(time.Minute * 60)

	if form.Now != 0 {
		now = time.UnixMilli(form.Now)
	}

	err = tj.financialService.CheckBalanceSnapshot(c, common.FINANCIAL_CODE_AUTO_YIELD, form.Type, form.Currencies, now)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	utils.ReData(c, nil)
}

// @Tags			tasks
// @Router			/tasks/financial/autoYield/snapshot [post]
// @Param			request			body		entities.AutoYieldSnapshotForm	false	"body"
func (tj *FinancialTaskJob) AutoYieldSnapshot(c *gin.Context) {

	now := time.Now().Truncate(time.Minute * 60)

	form := &entities.AutoYieldSnapshotForm{}

	err := c.ShouldBindJSON(&form)

	if err == nil {
		if form.Now != 0 {
			now = time.UnixMilli(form.Now)
		}
	}

	err = tj.financialService.BalanceSnapshot(c, common.FINANCIAL_CODE_AUTO_YIELD, now, form.Type, form.Currencies)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	utils.ReData(c, nil)
}

// @Tags			tasks
// @Param			request			body		entities.AutoYieldDistributeForm	false	"body"
// @Router			/tasks/financial/autoYield/distribute [post]
func (tj *FinancialTaskJob) AutoYieldDistribute(c *gin.Context) {

	now := time.Now().Truncate(time.Minute * 60)

	form := &entities.AutoYieldDistributeForm{}

	err := c.ShouldBindJSON(&form)

	if err == nil {
		if form.Now != 0 {
			now = time.UnixMilli(form.Now)
		}
	}

	err = tj.financialService.AutoYield(c, form.Type, form.Currencies, now)

	if err != nil {
		utils.ReError(c, err)
		return
	}

	utils.ReData(c, nil)
}
