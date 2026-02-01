package middleware

import (
	"api-server/lib"
	"api-server/services"
	"api-server/utils"
	"fmt"
	"shared-modules/common"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ApiAuthorityMiddleWare struct {
	authService *services.AuthService
	logger      lib.Logger
}

func NewApiAuthorityMiddleWare(authService *services.AuthService, logger lib.Logger) *ApiAuthorityMiddleWare {
	return &ApiAuthorityMiddleWare{
		authService: authService,
		logger:      logger,
	}
}

func (ah *ApiAuthorityMiddleWare) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {

		url := c.Request.URL.Path

		url = strings.TrimPrefix(url, "/api")

		key := c.Request.Header.Get("X-Token")

		err := ah.checkAPIAuth(c, url, key)
		if err != nil {
			utils.ReError(c, err)
			c.Abort()
			return
		}

		c.Next()
	}
}

func (ah *ApiAuthorityMiddleWare) checkAPIAuth(c *gin.Context, url string, key string) error {

	token, auths, err := ah.authService.CheckAPIAuthority(c, url, key)
	if err != nil {
		return err
	}

	rateLimit, err := ah.authService.RateLimit(c, token, auths)
	if rateLimit != nil {
		c.Writer.Header().Set(common.HEADER_X_RATELIMIT_LIMIT, strconv.Itoa(rateLimit.Limit))
		c.Writer.Header().Set(common.HEADER_X_RATELIMIT_REMAINING, strconv.Itoa(rateLimit.Remaining))
		c.Writer.Header().Set(common.HEADER_X_RATELIMIT_USED, strconv.Itoa(rateLimit.Used))
		c.Writer.Header().Set(common.HEADER_X_RATELIMIT_RESET, strconv.FormatInt(rateLimit.Reset.Unix(), 10))
	}
	if err != nil {
		return err
	}

	if token != nil {
		c.Set(common.HEADER_X_GROUP_IDS, strings.Trim(strings.Join(strings.Fields(fmt.Sprint(token.GroupIDs)), ","), "[]"))
		c.Set(common.HEADER_X_UID, token.UserID)
	}

	return nil
}
