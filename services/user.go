package services

import (
	authDao "api-server/dao/auth"
	userDao "api-server/dao/user"
	"api-server/lib"
	"context"
	"regexp"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type UserService struct {
	userDao      *userDao.UserDao
	userGroupDao *userDao.UserGroupDao
	tokenDao     *authDao.TokenDao
	logger       lib.Logger
}

func NewUserService(
	userDao *userDao.UserDao,
	userGroupDao *userDao.UserGroupDao,
	tokenDao *authDao.TokenDao,
	logger lib.Logger,
) *UserService {
	return &UserService{
		userDao:      userDao,
		userGroupDao: userGroupDao,
		tokenDao:     tokenDao,
		logger:       logger,
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
