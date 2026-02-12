package middlewares

import (
	"net/http"
	"strconv"

	"github.com/my-easy-vault-2026/shared-modules/common"

	authDao "github.com/my-easy-vault-2026/api-server/dao/auth"
	"github.com/my-easy-vault-2026/api-server/lib"
	"github.com/my-easy-vault-2026/api-server/services"

	"github.com/gin-gonic/gin"
)

type RateLimitMiddleWare struct {
	authService *services.AuthService
	logger      lib.Logger
	beBuilder   *lib.BEBuilder
	httpRes     *lib.HttpRes
}

func NewRateLimitMiddleWare(authService *services.AuthService, logger lib.Logger, beBuilder *lib.BEBuilder, httpRes *lib.HttpRes) *RateLimitMiddleWare {
	return &RateLimitMiddleWare{
		authService: authService,
		logger:      logger,
		beBuilder:   beBuilder,
		httpRes:     httpRes,
	}
}

func (rl *RateLimitMiddleWare) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {

		authsAny, ok := c.Get(common.CTX_KEY_AUTH_AUTHS)
		if !ok {
			rl.logger.Warn("no auths")
			rl.httpRes.ReError(c, http.StatusUnauthorized, rl.beBuilder.NewBusinessError(c, common.CODE_NO_PERMISSION))
			c.Abort()
			return
		}
		auths, ok := authsAny.([]*authDao.APIAuthority)
		if !ok {
			rl.logger.Warn("auths parse failed,", authsAny)
			rl.httpRes.ReError(c, http.StatusUnauthorized, rl.beBuilder.NewBusinessError(c, common.CODE_NO_PERMISSION))
			c.Abort()
			return
		}

		needToken := true
		for _, auth := range auths {
			if auth.Role == common.ROLE_GUEST {
				needToken = false
				break
			}
		}

		var token *authDao.Token
		if needToken {
			tokenAny, ok := c.Get(common.CTX_KEY_AUTH_TOKEN)
			if ok {
				token, ok = tokenAny.(*authDao.Token)
				if !ok {
					rl.logger.Warn("token parse failed,", tokenAny)
					rl.httpRes.ReError(c, http.StatusUnauthorized, rl.beBuilder.NewBusinessError(c, common.CODE_NO_PERMISSION))
					c.Abort()
					return
				}

			}
		}

		err := rl.checkRateLimit(c, token, auths)
		if err != nil {
			rl.httpRes.ReError(c, http.StatusTooManyRequests, err)
			c.Abort()
			return
		}

		c.Next()
	}
}

func (rl *RateLimitMiddleWare) checkRateLimit(c *gin.Context, token *authDao.Token, auths []*authDao.APIAuthority) error {

	rateLimit, err := rl.authService.RateLimit(c, token, auths)
	if rateLimit != nil {
		c.Writer.Header().Set(common.HEADER_X_RATELIMIT_LIMIT, strconv.Itoa(rateLimit.Limit))
		c.Writer.Header().Set(common.HEADER_X_RATELIMIT_REMAINING, strconv.Itoa(rateLimit.Remaining))
		c.Writer.Header().Set(common.HEADER_X_RATELIMIT_USED, strconv.Itoa(rateLimit.Used))
		c.Writer.Header().Set(common.HEADER_X_RATELIMIT_RESET, strconv.FormatInt(rateLimit.Reset.Unix(), 10))
	}
	if err != nil {
		return err
	}

	return nil
}
