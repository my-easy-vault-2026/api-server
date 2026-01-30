package services

import (
	accountDao "api-server/dao/account"
	authDao "api-server/dao/auth"
	cardDao "api-server/dao/card"
	systemDao "api-server/dao/system"
	userDao "api-server/dao/user"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/sha3"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

type AuthService struct {
	tokenDao        *authDao.TokenDao
	userDao         *userDao.UserDao
	userGroupDao    *userDao.UserGroupDao
	apiAuthorityDao *authDao.APIAuthorityDao
	parameterDao    *systemDao.ParameterDao
	userLoginLogDao *userDao.UserLoginLogDao
	categoryDao     *accountDao.CategoryDao
	assetDao        *accountDao.AssetDao
	mainCardDao     *cardDao.MainCardDao
	cardDao         *cardDao.CardDao
}

func NewAuthService() *AuthService {
	return &AuthService{
		tokenDao:        authDao.NewTokenDao(),
		userDao:         userDao.NewUserDao(),
		userGroupDao:    userDao.NewUserGroupDao(),
		apiAuthorityDao: authDao.NewAPIAuthorityDao(),
		parameterDao:    systemDao.NewParameterDao(),
		userLoginLogDao: userDao.NewUserLoginLogDao(),
		categoryDao:     accountDao.NewCategoryDao(),
		assetDao:        accountDao.NewAssetDao(),
		mainCardDao:     cardDao.NewMainCardDao(),
		cardDao:         cardDao.NewCardDao(),
	}
}

func (as *AuthService) LoginOrCreate(ctx context.Context, form *entities.LoginOrCreateForm) (*userDao.User, string, time.Time, bool, error) {
	isNewUser := false
	optLimitKey := utils.GetOptLimitKey(form.Email)
	val, err := utils.RDB.Get(ctx, optLimitKey).Int()
	if (err != redis.Nil) && (err == nil) {
		//TODO 未來系統參數直接從 redis 取用
		// 取回系統設定的 opt重試上限、
		limitTop, err := as.parameterDao.GetByKey(ctx, common.PARAMETER_KEY_OPT_LIMIT_TOP)
		if err != nil {
			logger.Warn("limitTop get failed,", err)
			return nil, "", time.Time{}, isNewUser, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		if limitTop == nil {
			logger.Warnf("limitTop no parameter: %s", common.PARAMETER_KEY_OPT_LIMIT_TOP)
			return nil, "", time.Time{}, isNewUser, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		if limitTop == nil {
			logger.Warnf("limitTop no parameter: %s", common.PARAMETER_KEY_OPT_LIMIT_TOP)
			return nil, "", time.Time{}, isNewUser, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		threshold, err := strconv.Atoi(limitTop.Value)
		if err != nil {
			logger.Warn("strconv failed,", err)
			return nil, "", time.Time{}, isNewUser, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		// 如果 key 存在，且值小於閾值，則進行 +1 操作
		newVal := val + 1
		if newVal > threshold {
			return nil, "", time.Time{}, isNewUser, utils.NewBusinessError(ctx, common.CODE_AUTH_OTP_RETRY_LIMIT_EXCEEDED)
		}
	}

	query := &userDao.UserQuery{}
	copier.Copy(query, form)
	query.PromotionCode = ""
	query.Role = common.ROLE_USER

	otpKey := utils.GetOTPRedisKey(common.MESSAGE_PURPOSE_USER_LOGIN_OR_REGISTER, common.NOTIFY_TYPE_EMAIL, form.OTPCode)

	getRet := utils.RDB.Get(
		ctx,
		otpKey,
	)

	if getRet.Err() != nil {
		if errors.Is(getRet.Err(), redis.Nil) {
			//累計五次鎖定10分鐘
			err := as.checkOptLimit(ctx, form)
			if err != nil {
				return nil, "", time.Time{}, isNewUser, err
			}
			return nil, "", time.Time{}, isNewUser, utils.NewBusinessError(ctx, common.CODE_AUTH_INCORRECT_EMAIL_OR_OTP)
		}
		logger.Warn("redis get failed", getRet.Err())
		return nil, "", time.Time{}, isNewUser, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if getRet.Val() != form.Email {
		//累計五次鎖定10分鐘
		err := as.checkOptLimit(ctx, form)
		if err != nil {
			return nil, "", time.Time{}, isNewUser, err
		}
		return nil, "", time.Time{}, isNewUser, utils.NewBusinessError(ctx, common.CODE_AUTH_INCORRECT_EMAIL_OR_OTP)
	}

	user, err := as.userDao.Get(ctx, query)
	if err != nil {
		logger.Warn("db get failed", err)
		return nil, "", time.Time{}, isNewUser, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	if user == nil {
		isNewUser = true
		logger.Infof("new user: %s", form.Email)

		user = &userDao.User{
			Email:        form.Email,
			Role:         common.ROLE_USER,
			KycLevel:     common.KYC_LEVEL_0,
			CoinfaceMain: common.COINFACE_MAIN_NO,
			Auto3DS:      common.AUTO_3DS_STATUS_DISABLED,
			AutoTopUp:    common.AUTO_TOP_UP_STATUS_DISABLED,
			ATMToggle:    common.ATM_TOGGLE_ENABLED,
		}

		users, err := as.userDao.ListByEmail(ctx, form.Email)
		if err != nil {
			logger.Warnf("ListByEmail: %v", err)
			return nil, "", time.Time{}, isNewUser, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		var mainUser *userDao.User
		for _, u := range users {
			if u.CoinfaceMain == common.COINFACE_MAIN_YES && u.KycLevel == common.KYC_LEVEL_3 {
				mainUser = u
				break
			}
		}

		if mainUser != nil {
			logger.Infof("user: [%s] already has kyc 3 on user [%d]", form.Email, mainUser.ID)
			var tErr error
			err = utils.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
				var ctx = context.WithValue(ctx, "db", tx)
				var rowsAffected int64
				rowsAffected, tErr = as.userDao.Update(ctx, &userDao.UserQuery{
					User: userDao.User{
						ID: user.ID,
					},
					Attrs: userDao.User{
						CountryCode: mainUser.CountryCode,
						PhoneNumber: mainUser.PhoneNumber,
						FirstName:   mainUser.FirstName,
						LastName:    mainUser.LastName,
						NationCode:  mainUser.NationCode,
						Gender:      mainUser.Gender,
						KycLevel:    common.KYC_LEVEL_3,
					},
				})
				if tErr != nil {
					logger.Warn("update kycLevel failed,", tErr)
					return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR).Wrap(tErr)
				}
				if rowsAffected == 0 {
					logger.Warn("update kycLevel failed, user id: [%d]", user.ID)
					return utils.NewBusinessError(ctx, common.CODE_DB_UPDATE_FAILED)
				}

				if rowsAffected == 0 {
					logger.Infof("update user limits failed, user_id: %d, limit_type: %d", user.ID, common.TRANSFER_LIMIT)
				}

				if tErr != nil {
					logger.Warn("update userLimits failed,", tErr)
					return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR).Wrap(tErr)
				}
				if rowsAffected == 0 {
					logger.Infof("update user limits failed, user_id: %d, limit_type: %d", user.ID, common.CP_EXPRESS_LIMIT)
				}

				return nil
			})
			if tErr != nil {
				logger.Warn("transaction failed,", tErr)
				return nil, "", time.Time{}, isNewUser, tErr
			}
			if err != nil {
				logger.Warn("transaction failed,", err)
				return nil, "", time.Time{}, isNewUser, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR).Wrap(err)
			}

			isNewUser = false
		}
	}

	// 確認用戶是否有邀請碼，如果沒有則設置邀請碼
	if user != nil {

		if user.Role != common.ROLE_USER {
			return nil, "", time.Time{}, isNewUser, utils.NewBusinessError(ctx, common.CODE_NO_PERMISSION)
		}

	}

	defer func() {
		delRet := utils.RDB.Del(
			ctx,
			otpKey,
		)
		if delRet.Err() != nil {
			logger.Warnf("redis failed to delete otp key: %s, %v", otpKey, delRet.Err())
		}
	}()

	key, expiredAt, err := as.GenerateAuthToken(ctx, *user, form)
	if err != nil {
		logger.Warnf("GenerateAuthToken: %s", err)
		return nil, "", time.Time{}, isNewUser, err
	}

	return user, key, expiredAt, isNewUser, nil
}

func (as *AuthService) GenerateAuthToken(ctx context.Context, user userDao.User, form *entities.LoginOrCreateForm) (string, time.Time, error) {

	key := utils.Md5String(form.Email + time.Now().String())
	wsToken := utils.Md5String("websocket" + form.Email + time.Now().String())
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(time.Hour * utils.Config.Auth.ExpireTime)
	token := &authDao.Token{
		UserID:     user.ID,
		GroupIDs:   make([]uint64, 0),
		MerchantID: user.MerchantID,
		HasPinCode: user.PinCode != "",
		Role:       common.ROLE_USER,
		WsToken:    wsToken,
		IssuedAt:   issuedAt,
		ExpiredAt:  expiredAt,
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
		time.Hour*utils.Config.Auth.ExpireTime,
		time.Hour*utils.Config.Auth.LoginDataExpireHours,
	)
	if err != nil {
		logger.Warn("set failed", err)
		return "", time.Time{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	wsKey := utils.GetWsTokenRedisKey(common.ROLE_USER, wsToken)
	err = utils.RDB.Set(ctx, wsKey, token.UserID, time.Hour*utils.Config.Auth.ExpireTime).Err()
	if err != nil {
		logger.Warn("save failed", err)
		return "", time.Time{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	// 設置 device & token， fido/removeDevice 同時需移除 token
	deviceTokenKey := utils.GetDeviceTokenRedisKey(user.ID, form.DeviceID)
	setRet := utils.RDB.Set(
		ctx,
		deviceTokenKey,
		utils.GetTokenRedisKey(key),
		time.Hour*utils.Config.Auth.ExpireTime,
	)

	if setRet.Err() != nil {
		logger.Warn("deviceTokenKey set failed", setRet.Err())
		return "", time.Time{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	//記錄登入資訊
	_, err = as.userLoginLogDao.Save(ctx, &userDao.UserLoginLog{
		ID:         utils.SnowFlakeID.Generate(),
		UserID:     user.ID,
		Platform:   form.Platform,
		DeviceName: form.DeviceName,
		DeviceID:   form.DeviceID,
		LoginIP:    form.Ip,
		AppVersion: form.AppVersion,
	})

	if err != nil {
		logger.Warn("user login log save failed", err)
	}

	return key, expiredAt, nil
}

func (as *AuthService) getUserLimitsDecimalParameter(ctx context.Context, key common.ParameterKey) (decimal.Decimal, error) {
	param, err := as.parameterDao.GetByKey(ctx, key)
	if err != nil {
		logger.Warn("get failed,", err)
		return decimal.Decimal{}, err
	}
	decimalValue, err := decimal.NewFromString(param.Value)
	if err != nil {
		logger.Warn("param parse failed,", err)
		return decimal.Decimal{}, err
	}
	return decimalValue, nil
}

func (as *AuthService) getUserLimitsIntParameter(ctx context.Context, key common.ParameterKey) (int, error) {
	param, err := as.parameterDao.GetByKey(ctx, key)
	if err != nil {
		logger.Warn("get failed,", err)
		return 0, err
	}
	intValue, err := strconv.Atoi(param.Value)
	if err != nil {
		logger.Warn("param parse failed,", err)
		return 0, err
	}
	return intValue, nil
}

func (as *AuthService) checkOptLimit(ctx context.Context, form *entities.LoginOrCreateForm) error {
	//TODO 未來系統參數直接從 redis 取用
	// 取回系統設定的 opt重試上限、opt重試時間
	limitTop, err := as.parameterDao.GetByKey(ctx, common.PARAMETER_KEY_OPT_LIMIT_TOP)
	if err != nil {
		logger.Warn("limitTop get failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if limitTop == nil {
		logger.Warnf("limitTop no parameter: %s", common.PARAMETER_KEY_OPT_LIMIT_TOP)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	limitInterval, err := as.parameterDao.GetByKey(ctx, common.PARAMETER_KEY_OPT_LIMIT_INTERVAL)
	if err != nil {
		logger.Warn("limitInterval get failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if limitInterval == nil {
		logger.Warnf("no parameter: %s", common.PARAMETER_KEY_OPT_LIMIT_INTERVAL)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	threshold, err := strconv.Atoi(limitTop.Value)
	if err != nil {
		logger.Warn("strconv failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	duration, err := strconv.Atoi(limitInterval.Value)
	if err != nil {
		logger.Warn("strconv failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	optLimitKey := utils.GetOptLimitKey(form.Email)
	val, err := utils.RDB.Get(ctx, optLimitKey).Int()
	if err == redis.Nil {
		// 如果 key 不存在，則進行初始設定
		err = utils.RDB.Set(ctx, optLimitKey, 1, time.Duration(duration)*time.Minute).Err()
		if err != nil {
			logger.Warn("Failed to set OptLimit initial value:", err)
			return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
	} else if err != nil {
		logger.Warn("Failed to get value:", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	} else {
		// 如果 key 存在，且值小於閾值，則進行 +1 操作
		newVal := val + 1
		if newVal > threshold {
			return utils.NewBusinessError(ctx, common.CODE_AUTH_OTP_RETRY_LIMIT_EXCEEDED)
		}
		// 如果次數等於設定的上限次數，則設定 10分鐘
		if newVal == threshold {
			logger.Infof("User: %s optLimit Value reached the threshold, stopping increment.", form.Email)
			// 取出鎖定時間
			limitLock, err := as.parameterDao.GetByKey(ctx, common.PARAMETER_KEY_OPT_LIMIT_LOCK_TIME)
			if err != nil {
				logger.Warn("get failed,", err)
				return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			}
			if limitLock == nil {
				logger.Warnf("no parameter: %s", common.PARAMETER_KEY_OPT_LIMIT_LOCK_TIME)
				return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			}
			lockTime, err := strconv.Atoi(limitLock.Value)
			if err != nil {
				logger.Warn("strconv failed,", err)
				return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			}
			// 重設鎖定時間
			err = utils.RDB.Set(ctx, optLimitKey, newVal, time.Duration(lockTime)*time.Minute).Err()
			if err != nil {
				logger.Warn("Failed to set OptLimit initial value:", err)
				return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			}

			return utils.NewBusinessError(ctx, common.CODE_AUTH_OTP_RETRY_LIMIT_EXCEEDED)
		}
		err = utils.RDB.Set(ctx, optLimitKey, newVal, time.Duration(duration)*time.Minute).Err()
		if err != nil {
			logger.Warn("Failed to increment value:", err)
			return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
	}
	return nil
}

func (as *AuthService) Logout(ctx context.Context, token, deviceID string, userID uint64) error {

	err := as.tokenDao.Remove(ctx, token)

	if err != nil {
		logger.Warn("remove failed", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	tokenKey := utils.GetFirebaseTokenKey(userID)
	err = utils.RDB.HDel(ctx, tokenKey, deviceID).Err()
	if err != nil {
		logger.Warn("remove firebase token failed", err)
	}

	deviceTokenKey := utils.GetDeviceTokenRedisKey(userID, deviceID)
	err = utils.RDB.Del(ctx, deviceTokenKey).Err()
	if err != nil {
		logger.Warn("remove device token failed", err)
	}

	/* josh: 因為token結構改變，調整一下這個部分 20240812
	// 根據 token 生成 tokenKey
	tokenKey := utils.GetTokenRedisKey(common.ROLE_USER, token)

	// 刪除 Redis 中的 tokenKey
	delRet := utils.RDB.Del(ctx, tokenKey)
	if delRet.Err() != nil {
		if errors.Is(delRet.Err(), redis.Nil) {
			return utils.NewBusinessError(ctx, common.CODE_NOT_LOGIN)
		}
		logger.Warnf("redis failed to delete token key: %s, %v", tokenKey, delRet.Err())
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	// 判斷是否真的刪除到資料，如果沒有 則刪除另一種 token
	if delRet.Val() == 0 {
		tokenKey := utils.GetTokenRedisKey(common.LOGIN_TYPE_USER_NO_PIN_CODE, token)
		// 刪除 Redis 中的 tokenKey
		delRet := utils.RDB.Del(ctx, tokenKey)
		if delRet.Err() != nil {
			if errors.Is(delRet.Err(), redis.Nil) {
				return utils.NewBusinessError(ctx, common.CODE_NOT_LOGIN)
			}
			logger.Warnf("redis failed to delete token key: %s, %v", tokenKey, delRet.Err())
			return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
	}
	*/

	return nil
}

func (as *AuthService) CheckAPIAuthority(ctx context.Context, url string, key string, deviceId string) (*authDao.Token, []*authDao.APIAuthority, error) {

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
			err = as.Logout(ctx, key, deviceId, token.UserID)
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

		switch authority.PinCodeRequired {
		case common.PINCODE_REQUIRED_YES:
			if !token.HasPinCode {
				possibleErr = utils.NewBusinessError(ctx, common.CODE_NO_PIN_CODE)
				continue
			}
		case common.PINCODE_REQUIRED_NO:
			if token.HasPinCode {
				continue
			}
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

		if authority.AdminLevel != 0 && token.Level < authority.AdminLevel {
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

func (as *AuthService) VerifyReapSign(ctx context.Context, bodyBytes []byte, timestamp int64, sign string) (bool, error) {
	if time.Unix(timestamp/1e3, timestamp%1e3).Add(time.Second * utils.Config.Reap.SignatureExpireSeconds).Before(time.Now()) {
		return false, utils.NewBusinessError(ctx, common.CODE_AUTH_REAP_SIGN_EXPIRED)
	}

	signBytes, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		logger.Warnf("decode failed: [%s][%v]", sign, err)
		return false, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	hash := sha3.New256()
	_, err = hash.Write(bodyBytes)
	if err != nil {
		logger.Warnf("hash failed: %v", err)
		return false, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	h := hash.Sum(nil)

	pubKey, err := base64.StdEncoding.DecodeString(utils.Config.Reap.PublicKey)
	if err != nil {
		logger.Warnf("decode failed: %v", err)
		return false, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	success, err := utils.RsaVerifySignWithSha256(h, signBytes, pubKey)
	if err != nil {
		logger.Warnf("verify failed: %v", err)
		return false, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return success, nil
}

func appendBody(params map[string]interface{}) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var bodyParts []string
	for _, k := range keys {
		v := params[k]
		switch vv := v.(type) {
		case []interface{}:
			if k == "events" {
				// 對每個 item 做 json.Marshal，以保證轉義與 Java 一致
				var strArray []string
				for _, item := range vv {
					strItem, ok := item.(string)
					if ok {
						// 使用 json.Marshal 來產出 "\"escaped\"" 格式
						escaped, _ := json.Marshal(strItem)
						strArray = append(strArray, string(escaped))
					}
				}
				bodyParts = append(bodyParts, fmt.Sprintf("%s=[%s]", k, strings.Join(strArray, ",")))
			} else {
				jsonBytes, _ := json.Marshal(vv)
				bodyParts = append(bodyParts, fmt.Sprintf("%s=%s", k, string(jsonBytes)))
			}
		default:
			bodyParts = append(bodyParts, fmt.Sprintf("%s=%v", k, v))
		}
	}
	return strings.Join(bodyParts, "&")
}

func (as *AuthService) VerifyPaycryptoSign(ctx context.Context, bodyBytes []byte, timestamp string, sign string) (bool, error) {
	var bodyData map[string]interface{}
	err := json.Unmarshal(bodyBytes, &bodyData)
	if err != nil {
		return false, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	dataForSign := appendBody(bodyData)

	logger.Infof("dataForSign: %s", dataForSign)

	// 取得 action
	actionVal, ok := bodyData["action"]
	if !ok {
		return false, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	action, ok := actionVal.(string)
	if !ok {
		return false, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	secretKey := utils.Config.Paycrypto.APISecret
	signature := utils.PaycryptoHmacSign(timestamp, action, secretKey, []byte(dataForSign))

	logger.Infof("signature: %s", signature)
	if sign == signature {
		return true, nil
	} else {
		return false, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
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

			pipe := utils.RDB.TxPipeline()
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
				pipe = utils.RDB.TxPipeline()
				pipe.Set(c, bucketKey, strconv.Itoa(a.Count), a.Window*time.Second*2)
				pipe.Set(c, dataKey, now.Format(time.RFC3339Nano), a.Window*time.Second*2)
				_, err := pipe.Exec(c)
				if err != nil {
					logger.Warnf("tx failed: %v", err)
					return nil, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR)
				}
				newGenAt = now
			case update:
				pipe = utils.RDB.TxPipeline()
				pipe.IncrBy(c, bucketKey, int64(gen))
				pipe.GetSet(c, dataKey, newGenAt.Format(time.RFC3339Nano))

				cmds, err := pipe.Exec(c)
				if err != nil && !errors.Is(err, redis.Nil) {
					logger.Warnf("tx failed: %v", err)
					return nil, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR)
				}
				if newTokenLeft, newBucketData := cmds[0].(*redis.IntCmd).Val(), cmds[1].(*redis.StringCmd).Val(); newBucketData != bucketData {
					logger.Warnf("concurrenct token incr. before: [%d],[%s]. after: [%d],[%s]", newTokenLeft-int64(gen), lastGenAt.Format(time.RFC3339Nano), newTokenLeft, newBucketData)

					pipe = utils.RDB.TxPipeline()
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
					pipe = utils.RDB.TxPipeline()
					pipe.Set(c, bucketKey, strconv.Itoa(a.Count), a.Window*time.Second*2)
					pipe.Set(c, dataKey, now.Format(time.RFC3339Nano), a.Window*time.Second*2)
					_, err := pipe.Exec(c)
					if err != nil {
						logger.Warnf("tx failed: %v", err)
						return nil, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR)
					}
				}
			}

			getRes := utils.RDB.Get(c, bucketKey)
			if getRes.Err() != nil {
				logger.Warnf("get failed: %v", getRes.Err())
				return nil, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR)
			}

			tokenLeft, err = strconv.Atoi(getRes.Val())
			if err != nil {
				logger.Warnf("parse failed: %v", err)
				utils.RDB.Del(c, bucketKey)
			}
			if tokenLeft <= 0 {
				rateLimits[i].Remaining = 0
				rateLimits[i].Used = a.Count
				rateLimits[i].Reset = newGenAt.Add(a.Window * time.Second)
				return rateLimits[i], utils.NewBusinessError(c, common.CODE_TOO_MANY_REQUEST)
			}

			decrRes := utils.RDB.Decr(c, bucketKey)
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

func (as *AuthService) PINUnlock(ctx context.Context, form *entities.PINUnlockForm, deviceID string, isPasskeyVerify bool, token string, userID uint64) error {

	locker := utils.NewLocker()
	if err := locker.WaitLock(
		ctx,
		utils.GetGlobalLockKey(common.LOCK_PURPOSE_PIN_UNLOCK, token),
		utils.Config.System.RedisDaoLockMicroseconds*time.Microsecond,
		utils.Config.System.LockWaitMicroseconds*time.Microsecond,
	); err != nil {
		logger.Warnf("lock failed: [%s], #v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_PIN_UNLOCK, token), err)
		return err
	}
	defer func() {
		if err := locker.UnLock(ctx); err != nil {
			logger.Warnf("unlock %s failed, %v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_PIN_UNLOCK, token), err)
		}
	}()

	user, err := as.userDao.GetByUserID(ctx, userID)
	if err != nil {
		logger.Warn("get failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	if user == nil {
		return utils.NewBusinessError(ctx, common.CODE_AUTH_UNLOCK_LIMIT_EXCEEDED)
	}

	redisKey := utils.GetUnlockTokenAttemptRedisKey(token)
	if !isPasskeyVerify {
		// 驗證 pincode
		if utils.CheckBcryptHash(form.PinCode, user.Salt, user.PinCode) {

			attempt := 0
			attemptStr, err := utils.RDB.Get(ctx, redisKey).Result()
			if err != nil && !errors.Is(err, redis.Nil) {
				return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			}
			if !errors.Is(err, redis.Nil) {
				attempt, err = strconv.Atoi(attemptStr)
				if err != nil {
					logger.Warnf("parse failed: %v", err)
					return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
				}
			}
			attempt++

			if attempt >= utils.Config.Auth.PINUnlockMaxAttempts {
				err = as.Logout(ctx, token, deviceID, user.ID)
				if err != nil {
					logger.Warnf("logout failed: %v", err)
				}
				msg := utils.NewBusinessError(ctx, common.CODE_AUTH_UNLOCK_LIMIT_EXCEEDED).Error()
				return utils.NewBusinessError(ctx, common.CODE_NOT_LOGIN, msg)
			}

			err = utils.RDB.Set(ctx, redisKey, strconv.Itoa(attempt), time.Hour*utils.Config.Auth.ExpireTime).Err()
			if err != nil {
				return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			}

			return utils.NewBusinessErrorWithArg(ctx, common.CODE_AUTH_UNLOCK_FAILED,
				map[common.BusinessErrorArg]string{
					common.BUSINESS_ERROR_ARG_ATTEMPT: strconv.Itoa(attempt),
				})
		}
	}

	err = utils.RDB.Del(ctx, redisKey).Err()
	if err != nil && !errors.Is(err, redis.Nil) {
		logger.Warn("remove device token failed, ", err)
	}
	return nil
}
