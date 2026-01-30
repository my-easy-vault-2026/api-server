package tasks

import (
	coinsdoDao "api-server/dao/coinsdo"
	"api-server/services"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"

	"github.com/gin-gonic/gin"
)

// 定时任务由gocron 服务通过定时调用http接口来实现 定时任务url前缀为/tasks/xxx
type RateTaskJob struct {
	cryptoCurrencyService *services.CryptoCurrencyService
	systemService         *services.SystemService
	accountService        *services.AccountService
}

func NewRateTaskJob() *RateTaskJob {
	return &RateTaskJob{
		cryptoCurrencyService: services.NewCryptoCurrencyService(),
		systemService:         services.NewSystemService(),
		accountService:        services.NewAccountService(),
	}
}

// @Summary		rate process task
// @Description	get rate and save.
// @Tags			tasks
// @Router			/tasks/rate/rateProcess [post]
func (tj *RateTaskJob) RateProcess(c *gin.Context) {
	// 撈取 系統設定的 數幣
	cryptos, err := tj.cryptoCurrencyService.GetCryptoCurrencies(c)
	if err != nil {
		logger.Warn(c, "cryptoCurrency get failed,", err)
		return
	}

	exist := make(map[string]bool, len(cryptos))
	for i := 0; i < len(cryptos); i++ {
		if _, ok := exist[cryptos[i].CurrencyName]; ok {
			cryptos = append(cryptos[:i], cryptos[i+1:]...)
			i--
			continue
		}
		exist[cryptos[i].CurrencyName] = true
	}

	points, err := tj.cryptoCurrencyService.ListByType(c, common.ASSET_TYPE_POINT)
	if err != nil {
		return
	}
	if len(points) > 0 {
		cryptos = append(cryptos, points...)
	}

	categories, err := tj.accountService.ListCategoryByUsage(c, &entities.ListCategoryByUsageForm{
		Usages: []common.CategoryUsage{
			common.CATEGORY_USAGE_QUOTE,
		}})

	quotes := make(map[string]bool, len(cryptos))
	for _, c := range categories {
		quotes[c.Name] = true
	}
	quoteCryptos := make([]*coinsdoDao.CryptoCurrency, 0, len(cryptos))
	for _, crypto := range cryptos {
		if quotes[crypto.CurrencyName] {
			quoteCryptos = append(quoteCryptos, crypto)
		}
	}

	tj.systemService.CurrencyRateProcess(c, quoteCryptos)

}

// @Summary		rate ping task
// @Description	test connectivity.
// @Tags			tasks
// @Router			/tasks/rate/ratePing [post]
func (tj *RateTaskJob) RatePing(c *gin.Context) {
	err := tj.systemService.CurrencyRatePingAndSwitch(c)
	if err != nil {
		logger.Errorf("rate task failed,", err)
	}
}
