package lib

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"shared-modules/common"

	"github.com/gin-gonic/gin"
)

type APPVersionLib struct {
	logger    Logger
	beBuilder *BEBuilder
}

func NewAPPVersionLib(logger Logger, beBuilder *BEBuilder) *APPVersionLib {
	return &APPVersionLib{
		logger:    logger,
		beBuilder: beBuilder,
	}
}

type APPVersion struct {
	Major int
	Minor int
	Patch int
}

type APPVersionRange struct {
	Min APPVersion
	Max APPVersion
}

type VRHandlerFunc struct {
	Range APPVersionRange
	Func  gin.HandlerFunc
}

type APPVersionRangeHandler struct {
	handlers    []VRHandlerFunc
	mu          sync.RWMutex
	defaultFunc gin.HandlerFunc
}

func (avl *APPVersionLib) NewVRFunc(min string, max string, f gin.HandlerFunc) VRHandlerFunc {
	minVer, err := avl.ParseAPPVersion(context.Background(), min)
	if err != nil {
		avl.logger.Fatalf("parse app version failed [%s]: [%v]", min, err)
	}
	maxVer, err := avl.ParseAPPVersion(context.Background(), max)
	if err != nil {
		avl.logger.Fatalf("parse app version failed [%s]: [%v]", max, err)
	}
	return VRHandlerFunc{
		Range: APPVersionRange{
			Min: minVer,
			Max: maxVer,
		},
		Func: f,
	}
}

func (avl *APPVersionLib) VRHandle(defaultFunc gin.HandlerFunc, handlers ...VRHandlerFunc) gin.HandlerFunc {

	handler := &APPVersionRangeHandler{
		handlers:    handlers[:],
		defaultFunc: defaultFunc,
	}

	sort.Slice(handler.handlers, func(i, j int) bool {
		return handler.handlers[i].Range.Max.Compare(handler.handlers[j].Range.Max) > 0
	})

	return func(c *gin.Context) {
		appVerStr := c.Request.Header.Get(common.HEADER_X_APP_VERSION)
		if appVerStr == "" {
			appVerStr = "1.0.0"
		}
		appVer, err := avl.ParseAPPVersion(c, appVerStr)
		if err != nil {
			avl.logger.Error("parse app version failed:", err)
			handler.defaultFunc(c)
			return
		}

		for _, h := range handler.handlers {
			if h.Range.Contains(appVer) {
				h.Func(c)
				return
			}
		}
		handler.defaultFunc(c)
	}
}

func (avl *APPVersionLib) ParseAPPVersion(ctx context.Context, v string) (APPVersion, error) {

	parts := strings.Split(strings.TrimPrefix(v, "v"), ".")
	if len(parts) != 3 {
		return APPVersion{}, avl.beBuilder.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return APPVersion{}, err
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return APPVersion{}, err
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return APPVersion{}, err
	}

	return APPVersion{major, minor, patch}, nil
}

// Compare 比較版本
// 返回 -1 如果 v < other
// 返回 0 如果 v == other
// 返回 1 如果 v > other
func (v APPVersion) Compare(other APPVersion) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}
	return 0
}

func (v APPVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (vr APPVersionRange) Contains(v APPVersion) bool {
	return v.Compare(vr.Min) >= 0 && v.Compare(vr.Max) <= 0
}
