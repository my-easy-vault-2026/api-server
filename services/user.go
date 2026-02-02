package services

import (
	authDao "api-server/dao/auth"
	userDao "api-server/dao/user"
	"api-server/lib"
	"context"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"
)

type UserService struct {
	userDao      *userDao.UserDao
	userGroupDao *userDao.UserGroupDao
	tokenDao     *authDao.TokenDao
	logger       lib.Logger
	beBuilder    *lib.BEBuilder
}

func NewUserService(
	userDao *userDao.UserDao,
	userGroupDao *userDao.UserGroupDao,
	tokenDao *authDao.TokenDao,
	logger lib.Logger,
	beBuilder *lib.BEBuilder,
) *UserService {
	return &UserService{
		userDao:      userDao,
		userGroupDao: userGroupDao,
		tokenDao:     tokenDao,
		logger:       logger,
		beBuilder:    beBuilder,
	}
}

func (us *UserService) GetInfo(ctx context.Context, userID uint64) (*userDao.User, []uint64, error) {

	user, err := us.userDao.Get(ctx, &userDao.UserQuery{
		User: userDao.User{
			ID: userID,
		},
	})

	logger.Debugf("user: [%#v]", user)

	if err != nil {
		logger.Warn("get failed,", err)
		return nil, nil, us.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	if user == nil {
		return nil, nil, us.beBuilder.NewBusinessError(ctx, common.CODE_USER_NO_SUCH_USER)
	}

	groups, err := us.userGroupDao.ListByUserID(ctx, userID)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, nil, us.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	groupIDs := make([]uint64, 0)
	for _, g := range groups {
		groupIDs = append(groupIDs, g.GroupID)
	}

	return user, groupIDs, nil
}

func (us *UserService) GetUserRole(ctx context.Context, userID uint64, role common.Role) (*userDao.User, error) {
	if userID == 0 || role == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	user, err := us.userDao.Get(ctx, &userDao.UserQuery{
		User: userDao.User{
			ID:   userID,
			Role: role,
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
