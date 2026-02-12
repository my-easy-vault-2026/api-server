package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/my-easy-vault-2026/shared-modules/common"

	"github.com/my-easy-vault-2026/api-server/infra"
	"github.com/my-easy-vault-2026/api-server/lib"
	"github.com/my-easy-vault-2026/shared-modules/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

const maxLogBody = 1 << 16 // 64KB

type TraceIdMiddleWare struct {
	logger          lib.Logger
	APIRouter       infra.Router `name:"api"`
	WebsocketRouter infra.Router `name:"websocket"`
}

type TraceIdMiddleWareParams struct {
	fx.In
	Logger          lib.Logger
	APIRouter       infra.Router `name:"api"`
	WebsocketRouter infra.Router `name:"websocket"`
}

func NewTraceIdMiddleWare(
	p TraceIdMiddleWareParams,
) *TraceIdMiddleWare {
	return &TraceIdMiddleWare{
		logger:          p.Logger,
		APIRouter:       p.APIRouter,
		WebsocketRouter: p.WebsocketRouter,
	}
}

func (tm *TraceIdMiddleWare) Setup() {
	tm.APIRouter.Use(tm.Handle())
	tm.WebsocketRouter.Use(tm.Handle())
}

func (tm *TraceIdMiddleWare) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {

		startTime := time.Now()
		c.Set("reqAt", startTime)

		if strings.HasPrefix(c.Request.URL.Path, "/swagger") {
			c.Next()
			return
		}

		traceId := c.GetHeader(common.HEADER_X_TRACE_ID)

		if traceId == "" {
			traceId = utils.Md5String(time.Now().String())
		}

		c.Set(common.CTX_KEY_TRACE_ID, traceId)

		body := c.Request.Body

		switch c.Request.Method {
		case http.MethodPost:
			bodyBytes, err := io.ReadAll(body)
			if err != nil {
				tm.logger.Errorf("path=%v,read body fail,%v", c.Request.URL.RequestURI(), err)
				c.Writer.WriteHeader(http.StatusInternalServerError)
				c.Abort()
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			if len(bodyBytes) > 65536 {
				bodyBytes = []byte(`{"msg":"TOO LARGE TO DISPLAY"}`)
			}

			tm.logger.Infof("app_version=%v,ip=%v,host=%v,path=%v,req=%v", c.GetHeader(common.HEADER_X_APP_VERSION), c.GetHeader(common.HEADER_X_REAL_IP), c.Request.Host, c.Request.URL.RequestURI(), sanitizeBody(bodyBytes))
		case http.MethodGet:
			tm.logger.Infof("app_version=%v,ip=%v,host=%v,path=%v", c.GetHeader(common.HEADER_X_APP_VERSION), c.GetHeader(common.HEADER_X_REAL_IP), c.Request.Host, c.Request.URL.RequestURI())
		}

		c.Next()
	}
}

func sanitizeBody(bodyBytes []byte) string {
	// 空 body
	if len(bodyBytes) == 0 {
		// logger.Info("Empty body received")
		return ""
	}

	sanitizeWords := []string{
		"pincode",
		"password",
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
