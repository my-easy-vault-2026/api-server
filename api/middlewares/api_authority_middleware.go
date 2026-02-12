package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/my-easy-vault-2026/shared-modules/common"

	"github.com/my-easy-vault-2026/api-server/lib"
	"github.com/my-easy-vault-2026/api-server/services"

	"github.com/gin-gonic/gin"
)

type ApiAuthorityMiddleWare struct {
	authService *services.AuthService
	logger      lib.Logger
	beBuilder   *lib.BEBuilder
	httpRes     *lib.HttpRes
}

func NewApiAuthorityMiddleWare(authService *services.AuthService, logger lib.Logger, beBuilder *lib.BEBuilder, httpRes *lib.HttpRes) *ApiAuthorityMiddleWare {
	return &ApiAuthorityMiddleWare{
		authService: authService,
		logger:      logger,
		beBuilder:   beBuilder,
		httpRes:     httpRes,
	}
}

func (ah *ApiAuthorityMiddleWare) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {

		url := c.Request.URL.Path

		url = strings.TrimPrefix(url, "/api")

		key := c.Request.Header.Get("X-Token")

		err := ah.checkAPIAuth(c, url, key)
		if err != nil {
			ah.httpRes.ReError(c, http.StatusUnauthorized, err)
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
	if token == nil {
		c.Set(common.CTX_KEY_AUTH_AUTHS, auths)
		return nil
	}

	ah.logger.Infof("check token, userId=%d", token.UserID)

	c.Set(common.CTX_KEY_AUTH_TOKEN, token)
	c.Set(common.CTX_KEY_AUTH_AUTHS, auths)

	if token != nil {
		c.Set(common.CTX_KEY_AUTH_GROUP, strings.Trim(strings.Join(strings.Fields(fmt.Sprint(token.GroupIDs)), ","), "[]"))
		c.Set(common.CTX_KEY_AUTH_UID, token.UserID)
		c.Set(common.CTX_KEY_AUTH_ROLE, token.Role)
	}

	return nil
}
