package tasks

import (
	"api-server/lib"
	"api-server/services"
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
