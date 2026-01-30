package main

import (
	"api-server/services"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime/debug"
	"shared-modules/common"
	"shared-modules/logger"
	"shared-modules/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var authServ *services.AuthService

func SetMiddleware() {
	authServ = services.NewAuthService()
}

func Recover(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {

			reqID, ok := utils.GetMDCValue("reqId")
			if ok {
				c.Writer.Header().Set("Request-Id", reqID)
			}

			path := c.Request.URL.RequestURI()
			file, line, fn := utils.FormatStackOneLineWithCode()
			logger.Errorf("panic recovered on [%s][%s],%v", path, fn, err)

			logger.Errorf("SEARCH_CODE:%s|%d", file, line)
			logger.Warnf(string(debug.Stack()))
			c.JSON(500, gin.H{
				"code": "10000",
				"msg":  "Internal server error",
			})
		}
	}()
	c.Next()
}

func HeaderInjector(c *gin.Context) {
	c.Set(common.HEADER_ACCEPT_LANGUAGED, c.GetHeader(common.HEADER_ACCEPT_LANGUAGED))
	c.Set(common.HEADER_X_REQUEST_ID, c.GetHeader(common.HEADER_X_REQUEST_ID))

	c.Next()
}

const maxLogBody = 1 << 16 // 64KB

func sanitizeBody(bodyBytes []byte) string {
	// 空 body
	if len(bodyBytes) == 0 {
		// logger.Info("Empty body received")
		return ""
	}

	sanitizeWords := []string{
		"pincode",
		"password",
		"panNumber",
		"activationCode",
		// "token",
		// "secret",
		// "access_key",
		// "apikey",
	}

	// 1) 嘗試 JSON
	var js any
	if err := json.Unmarshal(bodyBytes, &js); err == nil {
		red := redactJSON(js, sanitizeWords)
		out, _ := json.Marshal(red)
		return truncateString(string(out), maxLogBody)
	}

	// 2) 嘗試 form-urlencoded
	s := string(bodyBytes)
	if vals, err := url.ParseQuery(s); err == nil {
		// 遮罩 key 命中的值
		for k := range vals {
			if containsWord(k, sanitizeWords) {
				for i := range vals[k] {
					vals[k][i] = "[REDACTED]"
				}
			} else {
				// 值裡面有敏感字也遮
				for i := range vals[k] {
					if containsWord(vals[k][i], sanitizeWords) {
						vals[k][i] = "[REDACTED]"
					}
				}
			}
		}
		return truncateString(vals.Encode(), maxLogBody)
	}

	// 3) 純文字 fallback：有敏感字就整段遮，否則截斷回傳
	if containsWord(strings.ToLower(s), sanitizeWords) {
		return "[REDACTED]"
	}
	return truncateString(s, maxLogBody)
}

// 遞迴遮罩 JSON 的 key / string 值
func redactJSON(v any, words []string) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, vv := range t {
			if containsWord(k, words) {
				out[k] = "[REDACTED]"
				continue
			}
			out[k] = redactJSON(vv, words)
		}
		return out
	case []any:
		for i := range t {
			t[i] = redactJSON(t[i], words)
		}
		return t
	case string:
		if containsWord(t, words) {
			return "[REDACTED]"
		}
		return t
	default:
		return v
	}
}

func containsWord(s string, words []string) bool {
	ls := strings.ToLower(s)
	for _, w := range words {
		if w == "" {
			continue
		}
		if strings.Contains(ls, strings.ToLower(w)) {
			return true
		}
	}
	return false
}

func truncateString(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	// 以 rune 截斷避免破壞 UTF-8
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	var buf bytes.Buffer
	buf.WriteString(string(runes[:max]))
	buf.WriteString("…")
	return buf.String()
}

func Logger(c *gin.Context) {

	startTime := time.Now()
	c.Set("reqAt", startTime)

	if strings.HasPrefix(c.Request.URL.Path, "/swagger") || strings.HasPrefix(c.Request.URL.Path, "/public/upload") {
		c.Next()
		return
	}

	reqId := c.GetHeader(common.HEADER_X_REQUEST_ID)

	if reqId == "" {
		reqId = utils.Md5String(time.Now().String())
		c.Request.Header.Set(common.HEADER_X_REQUEST_ID, reqId)
	}

	userId := c.GetHeader("X-Uid")

	utils.SetMDCValue("reqId", reqId)

	body := c.Request.Body

	//if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodOptions {
	if c.Request.Method == http.MethodPost {
		bodyBytes, err := io.ReadAll(body)
		if err != nil {
			logger.Errorf("path=%v,read body fail,%v", c.Request.URL.RequestURI(), err)
			c.Writer.WriteHeader(http.StatusInternalServerError)
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		if len(bodyBytes) > 65536 {
			bodyBytes = []byte(`{"msg":"TOO LARGE TO DISPLAY"}`)
		}

		logger.Infof("user=%v,app_version=%v,ip=%v,host=%v,path=%v,req=%v", userId, c.GetHeader(common.HEADER_X_APP_VERSION), c.GetHeader(common.HEADER_X_REAL_IP), c.Request.Host, c.Request.URL.RequestURI(), sanitizeBody(bodyBytes))
	} else {
		logger.Infof("user=%v,app_version=%v,ip=%v,host=%v,path=%v", userId, c.GetHeader(common.HEADER_X_APP_VERSION), c.GetHeader(common.HEADER_X_REAL_IP), c.Request.Host, c.Request.URL.RequestURI())
	}

	c.Next()
}

var excludedUrl = [...]string{
	"/web/abc/abc",
}

func isExcludedUrl(url string) bool {

	for i := 0; i < len(excludedUrl); i++ {
		if utils.UrlFilter(excludedUrl[i], url) {
			return true
		}
	}
	return false
}

func CheckAuth(c *gin.Context) {

	url := c.Request.URL.Path

	url = strings.TrimPrefix(url, "/api")

	if isExcludedUrl(url) {
		c.Next()
		return
	}

	key := c.Request.Header.Get("X-Token")
	deviceId := c.Request.Header.Get(common.HEADER_X_DEVICE_ID)
	platform := c.Request.Header.Get(common.HEADER_X_PLATFORM)

	err := checkAPIAuth(c, url, key, deviceId)
	if err != nil {
		utils.ReError(c, err)
		c.Abort()
		return
	}

	c.Request.Header.Add(common.HEADER_X_DEVICE_ID, deviceId)
	c.Request.Header.Add(common.HEADER_X_PLATFORM, platform)
	c.Next()
}

func WsCheckAuth(c *gin.Context) {

	url := c.Request.URL.Path

	url = strings.TrimPrefix(url, "/api")

	if isExcludedUrl(url) {
		c.Next()
		return
	}

	c.Next()
}

func checkAPIAuth(c *gin.Context, url string, key string, deviceId string) error {

	token, auths, err := authServ.CheckAPIAuthority(c, url, key, deviceId)
	if err != nil {
		return err
	}

	rateLimit, err := authServ.RateLimit(c, token, auths)
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
		c.Request.Header.Add(common.HEADER_X_UID, strconv.FormatUint(token.UserID, 10))
		c.Request.Header.Add(common.HEADER_X_GROUP_IDS, strings.Trim(strings.Join(strings.Fields(fmt.Sprint(token.GroupIDs)), ","), "[]"))
		c.Request.Header.Add(common.HEADER_X_MERCHANT_ID, strconv.FormatUint(token.MerchantID, 10))
		c.Request.Header.Add(common.HEADER_X_ROLE, token.Role.String())
		c.Request.Header.Add(common.HEADER_X_LEVEL, token.Level.String())
		c.Set(common.HEADER_X_UID, token.UserID)
	}

	return nil
}
