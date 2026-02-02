package services

import (
	authDao "api-server/dao/auth"
	cardDao "api-server/dao/card"
	systemDao "api-server/dao/system"
	userDao "api-server/dao/user"
	"api-server/infra"
	"api-server/lib"
	"context"
	"errors"
	"math"
	"shared-modules/common"
	"shared-modules/logger"
	"shared-modules/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	tokenDao        *authDao.TokenDao
	userDao         *userDao.UserDao
	userGroupDao    *userDao.UserGroupDao
	apiAuthorityDao *authDao.APIAuthorityDao
	parameterDao    *systemDao.ParameterDao
	cardDao         *cardDao.CardDao
	redis           infra.Redis
	env             *lib.Env
	logger          lib.Logger
	beBuilder       *lib.BEBuilder
}

func NewAuthService(
	tokenDao *authDao.TokenDao,
	userDao *userDao.UserDao,
	userGroupDao *userDao.UserGroupDao,
	apiAuthorityDao *authDao.APIAuthorityDao,
	parameterDao *systemDao.ParameterDao,
	cardDao *cardDao.CardDao,
	redis infra.Redis,
	env *lib.Env,
	logger lib.Logger,
	beBuilder *lib.BEBuilder,
) *AuthService {
	return &AuthService{
		tokenDao:        tokenDao,
		userDao:         userDao,
		userGroupDao:    userGroupDao,
		apiAuthorityDao: apiAuthorityDao,
		parameterDao:    parameterDao,
		cardDao:         cardDao,
		redis:           redis,
		env:             env,
		logger:          logger,
		beBuilder:       beBuilder,
	}
}

func (as *AuthService) LoginOrCreate(ctx context.Context, email string, pinCode string) (*userDao.User, string, time.Time, error) {

	user, err := as.userDao.Get(ctx, &userDao.UserQuery{
		User: userDao.User{
			Email: email,
		},
	})
	if err != nil {
		logger.Warn("db get failed", err)
		return nil, "", time.Time{}, as.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	if user == nil {
		logger.Infof("new user: %s", email)

		user = &userDao.User{
			Email: email,
			Role:  common.ROLE_USER,
		}
		salt, err := utils.GenerateSalt(as.env.SaltLength)
		if err != nil {
			logger.Warn("generate salt failed,", err)
			return nil, "", time.Time{}, as.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}

		// 將密碼與 salt 組合
		saltedPassword := pinCode + salt
		// 使用 bcrypt 哈希組合後的密碼
		hashedPinCode, err := bcrypt.GenerateFromPassword([]byte(saltedPassword), bcrypt.DefaultCost)
		if err != nil {
			logger.Warn("hash password failed,", err)
			return nil, "", time.Time{}, as.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		user.Salt = salt
		user.PinCode = string(hashedPinCode)
		userID, err := as.userDao.Save(ctx, user)
		if err != nil {
			logger.Warn("save failed,", err)
			return nil, "", time.Time{}, as.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		logger.Infof("user created: %d", userID)
	}

	saltedPassword := pinCode + user.Salt
	err = bcrypt.CompareHashAndPassword([]byte(user.PinCode), []byte(saltedPassword))
	if err != nil {
		logger.Infof("invalid pin code for user: %s", email)
		return nil, "", time.Time{}, as.beBuilder.NewBusinessError(ctx, common.CODE_EMAIL_OR_PIN_CODE_INVALID)
	}

	key, expiredAt, err := as.GenerateAuthToken(ctx, user, email)
	if err != nil {
		logger.Warnf("GenerateAuthToken: %s", err)
		return nil, "", time.Time{}, err
	}

	return user, key, expiredAt, nil
}

func (as *AuthService) GenerateAuthToken(ctx context.Context, user *userDao.User, email string) (string, time.Time, error) {

	key := utils.Md5String(email + time.Now().String())
	wsToken := utils.Md5String("websocket" + email + time.Now().String())
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(time.Hour * as.env.TokenExpireTime)
	token := &authDao.Token{
		UserID:    user.ID,
		GroupIDs:  make([]uint64, 0),
		Role:      common.ROLE_USER,
		WsToken:   wsToken,
		IssuedAt:  issuedAt,
		ExpiredAt: expiredAt,
	}

	groups, err := as.userGroupDao.ListByUserID(ctx, user.ID)
	if err != nil {
		logger.Warn("get failed", err)
		return "", time.Time{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	for _, group := range groups {
		token.GroupIDs = append(token.GroupIDs, group.ID)
	}

	err = as.tokenDao.Save(ctx,
		key,
		token,
		as.env.TokenExpireTime,
		as.env.LoginDataExpireTime,
	)
	if err != nil {
		logger.Warn("set failed", err)
		return "", time.Time{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	wsKey := utils.GetWsTokenRedisKey(common.ROLE_USER, wsToken)
	err = as.redis.Set(ctx, wsKey, token.UserID, time.Hour*as.env.TokenExpireTime).Err()
	if err != nil {
		logger.Warn("save failed", err)
		return "", time.Time{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return key, expiredAt, nil
}

func (as *AuthService) Logout(ctx context.Context, token string, userID uint64) error {

	err := as.tokenDao.Remove(ctx, token)

	if err != nil {
		logger.Warn("remove failed", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return nil
}

func (as *AuthService) CheckAPIAuthority(ctx context.Context, url string, key string) (*authDao.Token, []*authDao.APIAuthority, error) {

	auths := make([]*authDao.APIAuthority, 0)

	token, err := as.tokenDao.Get(ctx, key)
	if err != nil {
		logger.Warn("get failed", err)
		return nil, nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if token != nil {
		if token.ExpiredAt.Before(time.Now()) {
			return nil, nil, utils.NewBusinessError(ctx, common.CODE_AUTH_TOKEN_EXPIRED)
		}
		user, err := as.userDao.GetByUserID(ctx, token.UserID)
		if err != nil {
			logger.Warnf("get failed %v", err)
			return nil, nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		if user == nil {
			logger.Warnf("user not exist")
			err = as.Logout(ctx, key, token.UserID)
			if err != nil {
				logger.Warnf("remove token failed. %v %v", token.UserID, err)
				return nil, nil, utils.NewBusinessError(ctx, common.CODE_NOT_LOGIN)
			}
			return nil, nil, utils.NewBusinessError(ctx, common.CODE_NOT_LOGIN)
		}
	}

	authorities, err := as.apiAuthorityDao.List(ctx)
	if err != nil {
		logger.Warn("get failed", err)
		return nil, nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	var passed bool
	var possibleErr error
	for _, authority := range authorities {
		index := strings.Index(authority.Endpoint, "*")
		if index < 0 {
			if authority.Endpoint != url {
				continue
			}
		} else {
			prefix := authority.Endpoint[:index]
			if !strings.HasPrefix(url, prefix) {
				continue
			}
		}

		if authority.Role == common.ROLE_GUEST {
			passed = true
			auths = append(auths, authority)
			continue
		}

		if token == nil {
			continue
		}

		if authority.Role != token.Role {
			continue
		}

		hasGroupID := false
		if authority.GroupID == 0 {
			hasGroupID = true
		}
		for _, gid := range token.GroupIDs {
			if gid == authority.GroupID {
				hasGroupID = true
				break
			}
		}
		if !hasGroupID {
			continue
		}

		passed = true
		auths = append(auths, authority)
	}

	if !passed {
		if possibleErr != nil {
			return nil, nil, possibleErr
		}
		if token == nil {
			return nil, nil, utils.NewBusinessError(ctx, common.CODE_NOT_LOGIN)
		}
		return nil, nil, utils.NewBusinessError(ctx, common.CODE_NO_PERMISSION)
	}

	return token, auths, nil
}

func (as *AuthService) RateLimit(c *gin.Context, token *authDao.Token, auths []*authDao.APIAuthority) (*common.RateLimit, error) {

	rateLimits := make([]*common.RateLimit, len(auths))
	for i, auth := range auths {
		rateLimits[i] = &common.RateLimit{
			Limit:     auth.Count,
			Remaining: math.MaxInt,
		}
	}

	limitFunc := func(key string) (*common.RateLimit, error) {

		for i, a := range auths {
			if a.Window == 0 {
				rateLimits[i] = nil
				continue
			}

			bucketKey := utils.GetTokenBucketRedisKey(c.Request.URL.RequestURI(), key)
			dataKey := utils.GetTokenBucketDataRedisKey(c.Request.URL.RequestURI(), key)

			pipe := as.redis.TxPipeline()
			pipe.Expire(c, bucketKey, a.Window*time.Second*2)
			pipe.Expire(c, dataKey, a.Window*time.Second*2)
			pipe.Get(c, bucketKey)
			pipe.Get(c, dataKey)

			cmds, err := pipe.Exec(c)
			if err != nil && !errors.Is(err, redis.Nil) {
				logger.Warnf("tx failed: %v", err)
				return nil, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR)
			}

			now := time.Now()
			lastGenAt, newGenAt := now, now
			var create, update bool
			bucketData := ""
			var tokenLeft, gen int

			if cmds[2].(*redis.StringCmd).Val() == "" {
				create = true
			} else if cmds[3].(*redis.StringCmd).Val() == "" {
				create = true
			} else {
				tokenLeft, err = strconv.Atoi(cmds[2].(*redis.StringCmd).Val())
				if err != nil {
					logger.Warnf("parse failed [%s]: %v", cmds[2].(*redis.StringCmd).Val(), err)
					create = true
				}
				bucketData = cmds[3].(*redis.StringCmd).Val()
				t, err := time.Parse(time.RFC3339Nano, cmds[3].(*redis.StringCmd).Val())
				if err != nil {
					logger.Warnf("parse failed [%s]: %v", cmds[3].(*redis.StringCmd).Val(), err)
					create = true
				}
				lastGenAt = t

				window := a.Window * 1e6 / time.Duration(a.Count) // us per req
				gen = int(now.Sub(lastGenAt) / time.Microsecond / window)
				newGenAt = lastGenAt.Add(window * time.Microsecond * time.Duration(gen))
				logger.Infof("last: %s, now: %s, dur: %ds, window: %d, gen: %d, newGenAt: %s", lastGenAt.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano), now.Sub(lastGenAt)/time.Second, window, gen, newGenAt.Format(time.RFC3339Nano))
				if gen > a.Count {
					create = true
				} else if gen > 0 {
					if tokenLeft+gen > a.Count {
						create = true
					} else {
						update = true
					}
				}
			}

			switch true {
			case create:
				pipe = as.redis.TxPipeline()
				pipe.Set(c, bucketKey, strconv.Itoa(a.Count), a.Window*time.Second*2)
				pipe.Set(c, dataKey, now.Format(time.RFC3339Nano), a.Window*time.Second*2)
				_, err := pipe.Exec(c)
				if err != nil {
					logger.Warnf("tx failed: %v", err)
					return nil, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR)
				}
				newGenAt = now
			case update:
				pipe = as.redis.TxPipeline()
				pipe.IncrBy(c, bucketKey, int64(gen))
				pipe.GetSet(c, dataKey, newGenAt.Format(time.RFC3339Nano))

				cmds, err := pipe.Exec(c)
				if err != nil && !errors.Is(err, redis.Nil) {
					logger.Warnf("tx failed: %v", err)
					return nil, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR)
				}
				if newTokenLeft, newBucketData := cmds[0].(*redis.IntCmd).Val(), cmds[1].(*redis.StringCmd).Val(); newBucketData != bucketData {
					logger.Warnf("concurrenct token incr. before: [%d],[%s]. after: [%d],[%s]", newTokenLeft-int64(gen), lastGenAt.Format(time.RFC3339Nano), newTokenLeft, newBucketData)

					pipe = as.redis.TxPipeline()
					pipe.DecrBy(c, bucketKey, int64(gen))
					pipe.Set(c, dataKey, newBucketData, a.Window*time.Second*2)
					cmds, err := pipe.Exec(c)
					if err != nil {
						logger.Warnf("tx failed: %v", err)
						return nil, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR)
					}
					if decrToken := cmds[0].(*redis.IntCmd).Val(); decrToken <= 0 {
						logger.Warnf("rate over limit: %s", bucketKey)
						rateLimits[i].Remaining = 0
						rateLimits[i].Used = a.Count
						t, err := time.Parse(time.RFC3339Nano, newBucketData)
						if err != nil {
							logger.Warnf("parse failed [%s]: %v", newBucketData, err)
							create = true
						}
						rateLimits[i].Reset = t.Add(a.Window * time.Second)
						return rateLimits[i], utils.NewBusinessError(c, common.CODE_TOO_MANY_REQUEST)
					}
				}
				if newTokenLeft, newBucketData := cmds[0].(*redis.IntCmd).Val(), cmds[1].(*redis.StringCmd).Val(); newTokenLeft >= int64(a.Count) {
					logger.Infof("token full. before: [%d],[%s]. after: [%d],[%s]", newTokenLeft-int64(gen), lastGenAt.Format(time.RFC3339Nano), newTokenLeft, newBucketData)
					pipe = as.redis.TxPipeline()
					pipe.Set(c, bucketKey, strconv.Itoa(a.Count), a.Window*time.Second*2)
					pipe.Set(c, dataKey, now.Format(time.RFC3339Nano), a.Window*time.Second*2)
					_, err := pipe.Exec(c)
					if err != nil {
						logger.Warnf("tx failed: %v", err)
						return nil, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR)
					}
				}
			}

			getRes := as.redis.Get(c, bucketKey)
			if getRes.Err() != nil {
				logger.Warnf("get failed: %v", getRes.Err())
				return nil, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR)
			}

			tokenLeft, err = strconv.Atoi(getRes.Val())
			if err != nil {
				logger.Warnf("parse failed: %v", err)
				as.redis.Del(c, bucketKey)
			}
			if tokenLeft <= 0 {
				rateLimits[i].Remaining = 0
				rateLimits[i].Used = a.Count
				rateLimits[i].Reset = newGenAt.Add(a.Window * time.Second)
				return rateLimits[i], utils.NewBusinessError(c, common.CODE_TOO_MANY_REQUEST)
			}

			decrRes := as.redis.Decr(c, bucketKey)
			if decrRes.Err() != nil {
				logger.Warnf("decr failed: %v", decrRes.Err())
				return nil, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR)
			}

			rateLimits[i].Remaining = int(decrRes.Val())
			rateLimits[i].Used = a.Count - int(decrRes.Val())
			rateLimits[i].Reset = newGenAt.Add(a.Window * time.Duration(rateLimits[i].Used) * time.Microsecond)
		}

		minLimit := &common.RateLimit{Remaining: math.MaxInt}
		for _, limit := range rateLimits {
			if limit != nil && limit.Remaining < minLimit.Remaining {
				minLimit = limit
			}
		}
		if minLimit.Remaining != math.MaxInt {
			return minLimit, nil
		}
		return nil, nil
	}

	minLimit1, err := limitFunc(c.GetHeader(common.HEADER_X_REAL_IP))
	if err != nil {
		return minLimit1, err
	}

	if token == nil {
		return minLimit1, err
	}

	minLimit2, err := limitFunc(strconv.FormatUint(token.UserID, 10))
	if err != nil {
		return minLimit2, err
	}

	switch true {
	case minLimit1 == nil && minLimit2 != nil:
		return minLimit2, nil
	case minLimit1 != nil && minLimit2 == nil:
		return minLimit1, nil
	case minLimit1 == nil && minLimit2 == nil:
		return nil, nil
	case minLimit1.Remaining < minLimit2.Remaining:
		return minLimit1, nil
	default:
		return minLimit2, nil
	}
}
