package services

import (
	accountDao "api-server/dao/account"
	authDao "api-server/dao/auth"
	cardDao "api-server/dao/card"
	systemDao "api-server/dao/system"
	userDao "api-server/dao/user"
	"context"
	"errors"
	"regexp"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type UserService struct {
	userDao       *userDao.UserDao
	deleteUserDao *userDao.DeleteUserDao
	assetDao      *accountDao.AssetDao
	userGroupDao  *userDao.UserGroupDao
	tokenDao      *authDao.TokenDao
	parameterDao  *systemDao.ParameterDao
	cardDao       *cardDao.DeleteCardDao
}

func NewUserService() *UserService {

	return &UserService{
		userDao:       userDao.NewUserDao(),
		assetDao:      accountDao.NewAssetDao(),
		deleteUserDao: userDao.NewDeleteUserDao(),
		userGroupDao:  userDao.NewUserGroupDao(),
		tokenDao:      authDao.NewTokenDao(),
		parameterDao:  systemDao.NewParameterDao(),
	}
}

func (us *UserService) SetPinCode(ctx context.Context, form *entities.SetPinCodeForm, key string, userID uint64) error {

	user, err := us.userDao.Get(ctx, &userDao.UserQuery{
		User: userDao.User{
			ID: userID,
		},
	})

	if err != nil {
		logger.Warn("db get failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	if user == nil {
		return utils.NewBusinessError(ctx, common.CODE_USER_NO_SUCH_USER)
	}

	if user.PinCode != "" {
		return utils.NewBusinessError(ctx, common.CODE_USER_ALREADY_HAS_PIN_CODE)
	}

	salt, err := utils.GenerateSalt(utils.Config.System.SaltLength)
	if err != nil {
		logger.Warn("generate salt failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	hashedPinCode, err := utils.BcryptHash(form.PinCode, salt)

	if err != nil {
		logger.Warn("hash password failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	userQuery := &userDao.UserQuery{
		User: userDao.User{
			ID: userID,
		},
		Attrs: userDao.User{
			PinCode: hashedPinCode,
			Salt:    salt,
		},
	}

	if rowsAffected, err := us.userDao.Update(ctx, userQuery); err != nil {
		logger.Warn("db update failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	} else if rowsAffected != 1 {
		logger.Warnf("db update rows: [%v]", rowsAffected)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	groups, err := us.userGroupDao.ListByUserID(ctx, userID)
	if err != nil {
		logger.Warn("get failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	groupIDs := make([]uint64, 0)
	for _, g := range groups {
		groupIDs = append(groupIDs, g.GroupID)
	}
	now := time.Now()

	err = us.tokenDao.Save(ctx, key, &authDao.Token{
		UserID:     userID,
		GroupIDs:   groupIDs,
		MerchantID: user.MerchantID,
		HasPinCode: true,
		Role:       user.Role,
		IssuedAt:   now,
		ExpiredAt:  now.Add(time.Hour * utils.Config.Auth.ExpireTime),
	},
		time.Hour*utils.Config.Auth.ExpireTime,
		time.Hour*utils.Config.Auth.LoginDataExpireHours,
	)
	if err != nil {
		logger.Warn("save failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return nil
}

func (us *UserService) ForgotPinCode(ctx context.Context, form *entities.ForgotPinCodeForm, userID uint64) error {
	userIDString := ctx.(*gin.Context).Request.Header.Get(common.HEADER_X_UID)
	if userIDString == "" {
		logger.Error("no X-Uid")
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	userID, err := strconv.ParseUint(userIDString, 10, 64)
	if err != nil {
		logger.Error("X-Uid parse failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	user, err := us.userDao.Get(ctx, &userDao.UserQuery{
		User: userDao.User{
			ID: userID,
		},
	})

	if err != nil {
		logger.Warn("get failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	if user == nil {
		return utils.NewBusinessError(ctx, common.CODE_USER_NO_SUCH_USER)
	}

	otpKey := utils.GetOTPRedisKey(common.MESSAGE_PURPOSE_FORGET_PIN_CODE, common.NOTIFY_TYPE_EMAIL, form.OTPCode)

	getRet := utils.RDB.Get(
		ctx,
		otpKey,
	)

	if getRet.Err() != nil {
		if errors.Is(getRet.Err(), redis.Nil) {
			return utils.NewBusinessError(ctx, common.CODE_USER_INCORRECT_OTP)
		}
		logger.Warn("redis get failed", getRet.Err())
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	if getRet.Val() != user.Email {
		return utils.NewBusinessError(ctx, common.CODE_USER_INCORRECT_OTP)
	}

	count := make(map[rune]int)
	for _, digit := range form.NewPinCode {
		count[digit]++
		if count[digit] >= 3 {
			return utils.NewBusinessError(ctx, common.CODE_REAP_MORE_THAN_TWO_IDENTICAL_DIGITS)
		}
	}
	sequencePattern := "(?:012|123|234|345|456|567|678|789|987|876|765|654|543|432|321|210)"
	re := regexp.MustCompile(sequencePattern)
	if re.MatchString(form.NewPinCode) {
		return utils.NewBusinessError(ctx, common.CODE_REAP_MORE_THAN_TWO_SEQUENTAIL_DIGITS)
	}

	salt, err := utils.GenerateSalt(utils.Config.System.SaltLength)
	if err != nil {
		logger.Warn("generate salt failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	userNewHashPincode, err := utils.BcryptHash(form.NewPinCode, salt)
	if err != nil {
		logger.Warn("hash failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	if rowsAffected, err := us.userDao.Update(ctx, &userDao.UserQuery{
		User: userDao.User{
			ID:      userID,
			PinCode: user.PinCode,
		},
		Attrs: userDao.User{
			PinCode: userNewHashPincode,
			Salt:    salt,
		},
	}); err != nil {
		logger.Warn("update failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	} else if rowsAffected != 1 {
		logger.Warnf("update rows: [%v]", rowsAffected)
		return utils.NewBusinessError(ctx, common.CODE_USER_PIN_CODE_CONFIRMATION_FAILED) // 併發更新密碼
	}

	return nil
}

func (us *UserService) ResetPinCode(ctx context.Context, form *entities.ResetPinCodeForm) error {
	userIDString := ctx.(*gin.Context).Request.Header.Get(common.HEADER_X_UID)
	if userIDString == "" {
		logger.Error("no X-Uid")
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	userID, err := strconv.ParseUint(userIDString, 10, 64)
	if err != nil {
		logger.Error("X-Uid parse failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	user, err := us.userDao.Get(ctx, &userDao.UserQuery{
		User: userDao.User{
			ID: userID,
		},
	})

	if err != nil {
		logger.Warn("get failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	if user == nil {
		return utils.NewBusinessError(ctx, common.CODE_USER_NO_SUCH_USER)
	}

	if utils.CheckBcryptHash(form.OldPinCode, user.Salt, user.PinCode) {
		return utils.NewBusinessError(ctx, common.CODE_USER_PIN_CODE_CONFIRMATION_FAILED)
	}

	sequencePattern := "(?:012|123|234|345|456|567|678|789|987|876|765|654|543|432|321|210)"
	count := make(map[rune]int)
	for _, digit := range form.NewPinCode {
		count[digit]++
		if count[digit] >= 3 {
			return utils.NewBusinessError(ctx, common.CODE_REAP_MORE_THAN_TWO_IDENTICAL_DIGITS)
		}
	}
	re := regexp.MustCompile(sequencePattern)
	if re.MatchString(form.NewPinCode) {
		return utils.NewBusinessError(ctx, common.CODE_REAP_MORE_THAN_TWO_SEQUENTAIL_DIGITS)
	}

	salt, err := utils.GenerateSalt(utils.Config.System.SaltLength)
	if err != nil {
		logger.Warn("generate salt failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	userNewHashPincode, err := utils.BcryptHash(form.NewPinCode, salt)
	if err != nil {
		logger.Warn("hash failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	if rowsAffected, err := us.userDao.Update(ctx, &userDao.UserQuery{
		User: userDao.User{
			ID:      userID,
			PinCode: user.PinCode,
		},
		Attrs: userDao.User{
			PinCode: userNewHashPincode,
			Salt:    salt,
		},
	}); err != nil {
		logger.Warn("update failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	} else if rowsAffected != 1 {
		logger.Warnf("update rows: [%v]", rowsAffected)
		return utils.NewBusinessError(ctx, common.CODE_USER_PIN_CODE_CONFIRMATION_FAILED) // 併發更新密碼
	}

	return nil
}

func (us *UserService) GetInfo(ctx context.Context, userID uint64, form *entities.GetInfoForm) (*userDao.User, []uint64, error) {

	user, err := us.userDao.Get(ctx, &userDao.UserQuery{
		User: userDao.User{
			ID: userID,
		},
	})

	logger.Debugf("user: [%#v]", user)

	if err != nil {
		logger.Warn("get failed,", err)
		return nil, nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	if user == nil {
		return nil, nil, utils.NewBusinessError(ctx, common.CODE_USER_NO_SUCH_USER)
	}

	groups, err := us.userGroupDao.ListByUserID(ctx, userID)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	groupIDs := make([]uint64, 0)
	for _, g := range groups {
		groupIDs = append(groupIDs, g.GroupID)
	}

	return user, groupIDs, nil
}

func (us *UserService) GetUserRole(ctx context.Context, form *entities.GetUserForm) (*userDao.User, error) {
	if form.ID == 0 && (form.Email == "" || form.Role == 0) {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	user, err := us.userDao.Get(ctx, &userDao.UserQuery{
		User: userDao.User{
			ID:    form.ID,
			Email: form.Email,
			Role:  form.Role,
		},
	})

	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	if user == nil {
		return nil, utils.NewBusinessError(ctx, common.CODE_USER_NO_SUCH_USER)
	}

	return user, nil
}
func (us *UserService) ListByEmailRole(ctx context.Context, email string, role common.Role) ([]*userDao.User, error) {
	users, err := us.userDao.ListByEmailRole(ctx, email, role)
	if err != nil {
		logger.Warn("get user info fail,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR).Wrap(err)
	}

	return users, nil
}

func (us *UserService) ListUsers(ctx context.Context, form *entities.ListUsersForm) ([]*userDao.User, error) {

	users, err := us.userDao.ListByUserIDIn(ctx, form.UserIDs)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return users, nil
}

func (us *UserService) SavePhoneNumber(ctx context.Context, form entities.SavePhoneNumberForm) (bool, error) {
	userIDString := ctx.(*gin.Context).Request.Header.Get(common.HEADER_X_UID)

	countryCode, err := strconv.Atoi(form.CountryCode)
	if err != nil {
		logger.Warn("country code should be number, ", err)
		return false, utils.NewBusinessError(ctx, common.CODE_COUNTRY_CODE_SHOULD_BE_NUMBER)
	}

	userID, err := strconv.ParseUint(userIDString, 10, 64)
	if err != nil {
		logger.Error("X-Uid parse failed,", err)
		return false, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	user, err := us.userDao.Get(ctx, &userDao.UserQuery{
		User: userDao.User{
			ID: userID,
		},
	})

	if err != nil {
		logger.Warn("get failed,", err)
		return false, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	if user == nil {
		return false, utils.NewBusinessError(ctx, common.CODE_USER_NO_SUCH_USER)
	}

	_, err = us.userDao.Update(ctx, &userDao.UserQuery{
		User: userDao.User{
			ID: userID,
		},
		Attrs: userDao.User{
			CountryCode: countryCode,
			PhoneNumber: form.PhoneNumber,
		},
	})
	if err != nil {
		logger.Warn("get failed,", err)
		return false, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return true, nil
}

func (us *UserService) DeleteAccount(ctx context.Context, form *entities.DeleteAccountForm, userID uint64) error {

	user, err := us.userDao.Get(ctx, &userDao.UserQuery{
		User: userDao.User{
			ID: userID,
		},
	})
	if err != nil {
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if user == nil {
		return utils.NewBusinessError(ctx, common.CODE_CARD_INCORRECT_PIN_CODE)
	}

	// 驗證 pincode
	if utils.CheckBcryptHash(form.PinCode, user.Salt, user.PinCode) {
		return utils.NewBusinessError(ctx, common.CODE_USER_PIN_CODE_CONFIRMATION_FAILED)
	}

	users, err := us.userDao.ListByEmail(ctx, user.Email)
	if err != nil {
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if len(users) > 1 {
		return utils.NewBusinessError(ctx, common.CODE_ACCOUNT_INVALID_TARGET)
	}

	// 檢核面額，大於 10usd 不能刪除
	assets, err := us.assetDao.GetByUserIDTypesIn(ctx, user.ID, []common.AssetType{common.ASSET_TYPE_CRYPTO, common.ASSET_TYPE_CARD_PRODUCT, common.ASSET_TYPE_AUTO_YIELD})
	totalAmount := decimal.NewFromFloat(0.0)
	for _, asset := range assets {
		// 匯率處理
		exchangePrice := decimal.NewFromInt(1)
		rates := make([]*utils.ExchangeRate, 0, 2)
		rates, err = utils.ListExchangeRate(ctx, asset.Currency, []common.Currency{common.CURRENCY_USDT})
		if err != nil {
			logger.Warn("get exchange rate failed,", err)
			return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		for _, rate := range rates {
			if rate.QuoteCurrency == common.CURRENCY_USDT {
				exchangePrice = rate.Rate
			}
		}
		totalAmount = asset.Amount.Add(asset.FreezedAmount).Mul(exchangePrice)

		if totalAmount.GreaterThan(decimal.NewFromFloat(10.0)) {
			return utils.NewBusinessError(ctx, common.CODE_USER_UNABLE_DEACTIVATE_BY_REMAIN_ASSETS)
		}
	}

	err = utils.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		var c = context.WithValue(ctx, "db", tx)
		// 刪除 用戶資料 轉移到另一張表
		var deleteUser userDao.DeleteUser
		copier.Copy(&deleteUser, &user)
		_, err = us.deleteUserDao.Save(c, &deleteUser)

		if err != nil {
			return utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR)
		}

		err = us.userDao.DeleteByID(c, user.ID)
		if err != nil {
			return utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR)
		}

		return nil
	})

	return nil
}

func (us *UserService) CheckKYCLevel(ctx context.Context, userId uint64) (user *userDao.User, err error) {

	user, err = us.userDao.GetByUserID(ctx, userId)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if user == nil {
		return nil, utils.NewBusinessError(ctx, common.CODE_CARD_INCORRECT_PIN_CODE)
	}

	if user.KycLevel == common.KYC_LEVEL_0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_CARD_KYC_LEVEL_INSUFFICIENT)
	}
	return user, err
}

func (us *UserService) GetEmailByUserID(ctx context.Context, userID uint64) (string, error) {

	user, err := us.userDao.GetByUserID(ctx, userID)
	if err != nil {
		logger.Warn("get failed,", err)
		return "", utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return user.Email, nil
}

func (us *UserService) SaveLanguage(ctx context.Context, form entities.SaveLanguageForm) (bool, error) {

	_, err := us.userDao.Update(ctx, &userDao.UserQuery{
		User: userDao.User{
			ID: form.UserID,
		},
		Attrs: userDao.User{
			Language: form.Language,
		},
	})
	if err != nil {
		logger.Warn("get failed,", err)
		return false, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return true, nil
}

func (us *UserService) GetAssetByUserIDCurrency(ctx context.Context, userID uint64, currency common.Currency) (user *accountDao.Asset, err error) {

	asset, err := us.assetDao.GetByUserIDCurrency(ctx, userID, currency)
	if err != nil {
		logger.Warn("get asset failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if asset == nil {
		return nil, utils.NewBusinessError(ctx, common.CODE_CARD_INCORRECT_PIN_CODE)
	}

	return asset, err
}
