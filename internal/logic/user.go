package logic

import (
	"context"
	"errors"
	"fmt"
	"mp/internal/errno"
	"mp/internal/model"
	"mp/pkg/logx"

	"mp/internal/domain"
	"mp/pkg/encrypt"
)

type User interface {
	Login(ctx context.Context, req *domain.LoginReq) (resp *domain.LoginResp, err error)
	Register(ctx context.Context, req *domain.RegisterReq) (resp *domain.RegisterResp, err error)

	Info(ctx context.Context, req int64) (resp *domain.User, err error)
	Create(ctx context.Context, req *domain.User) (err error)
	List(ctx context.Context, d *domain.UserListReq) (*domain.UserListResp, error)

	DeleteSelf(ctx context.Context, uid int64) error
	AdminDelete(ctx context.Context, adminID, targetUID int64) error
	delete(ctx context.Context, req int64) (err error)

	UpPassword(ctx context.Context, uid int64, req *domain.UpPasswordReq) (err error)
}

type user struct {
	usersModel model.UsersModel
}

func NewUser(usersModel model.UsersModel) User {
	return &user{usersModel: usersModel}
}

func (l *user) Login(ctx context.Context, req *domain.LoginReq) (resp *domain.LoginResp, err error) {
	userEntity, err := l.usersModel.FindByNameOrPhone(req.Username)
	if err != nil {
		logx.Errors(ctx, "user", "login_failed", logx.Fields{
			"username": req.Username,
			"reason":   err.Error(),
		})
		return nil, err
	}

	if !encrypt.VaildPasswordHash(req.Password, userEntity.Password) {
		logx.Infos(ctx, "user", "login_failed", logx.Fields{
			"stage":    "password_check",
			"username": req.Username,
		})
		return nil, errno.ErrPasswordInvalid
	}

	logx.Infos(ctx, "user", "login_success", logx.Fields{
		"user_id": userEntity.Id,
	})

	return &domain.LoginResp{
		Id:     userEntity.Id,
		Name:   userEntity.Name,
		Status: int64(userEntity.Status),
	}, nil
}

func (l *user) Register(ctx context.Context, req *domain.RegisterReq) (*domain.RegisterResp, error) {
	userEntity, err := l.usersModel.FindByName(req.Name)
	if err != nil {
		return nil, err
	}
	if userEntity != nil {
		return nil, errno.ErrUserAlreadyExists
	}

	if req.Password != req.Password2 {
		return nil, errno.ErrPasswordMismatch
	}

	newUser := &domain.User{
		Name:     req.Name,
		Phone:    req.Phone,
		Password: req.Password,
		Status:   int64(model.UserStatusNormal),
		IsSystem: 0,
	}

	if err := l.createUser(ctx, newUser, "register"); err != nil {
		logx.Errors(ctx, "user", "register_failed", logx.Fields{
			"stage": "create_user",
			"name":  req.Name,
			"error": err.Error(),
		})
		return nil, err
	}

	logx.Infos(ctx, "user", "register_success", logx.Fields{
		"user_id": newUser.Id,
	})

	return &domain.RegisterResp{
		Id:     newUser.Id,
		Name:   newUser.Name,
		Status: int(newUser.Status),
	}, nil
}

func (l *user) Create(ctx context.Context, req *domain.User) error {
	userEntity, err := l.usersModel.FindByName(req.Name)
	if err != nil {
		logx.Errors(ctx, "admin", "admin_create_user_failed", logx.Fields{
			"stage": "check_name",
			"name":  req.Name,
			"error": err.Error(),
		})
		return err
	}

	if userEntity != nil {
		return errno.ErrUserAlreadyExists
	}

	if err := l.createUser(ctx, req, "admin"); err != nil {
		logx.Errors(ctx, "admin", "admin_create_user_failed", logx.Fields{
			"stage": "create_user",
			"name":  req.Name,
			"error": err.Error(),
		})
		return err
	}

	logx.Infos(ctx, "admin", "admin_create_user_success", logx.Fields{
		"name": req.Name,
	})

	return nil
}

func (l *user) createUser(ctx context.Context, req *domain.User, from string) error {
	passwordHash, err := encrypt.GenPasswordHash([]byte(req.Password))
	if err != nil {
		return fmt.Errorf("gen password hash failed: %w", err)
	}

	if err := l.usersModel.Insert(ctx, &model.Users{
		Name:     req.Name,
		Phone:    req.Phone,
		Password: string(passwordHash),
		Status:   model.UserStatus(req.Status),
		IsSystem: func() int64 {
			if from == "admin" {
				return 1
			}
			return 0
		}(),
	}); err != nil {
		// 一律视为系统异常
		return fmt.Errorf("insert user failed: %w", err)
	}

	return nil
}

func (l *user) Info(ctx context.Context, req int64) (resp *domain.User, err error) {
	user, err := l.usersModel.FindOne(ctx, req)
	if err != nil {
		return nil, err
	}
	return user.ToDomainUser(), nil
}

func (l *user) UpPassword(ctx context.Context, uid int64, req *domain.UpPasswordReq) error {
	userEntity, err := l.usersModel.FindOne(ctx, uid)
	if err != nil {
		return err
	}

	ok := encrypt.VaildPasswordHash(
		userEntity.Password,
		req.OldPwd,
	)
	if !ok {
		logx.Infos(ctx, "user", "up_password_failed", logx.Fields{
			"stage": "password_check",
			"uid":   uid,
		})
		return errno.ErrPasswordInvalid
	}

	if req.NewPwd == "" {
		return errno.ErrPasswordEmpty
	}
	if req.NewPwd == req.OldPwd {
		return errno.ErrPasswordSame
	}

	newHash, err := encrypt.GenPasswordHash([]byte(req.NewPwd))
	if err != nil {
		return err
	}

	userEntity.Password = string(newHash)

	return l.usersModel.Update(ctx, userEntity)
}

func (l *user) DeleteSelf(ctx context.Context, uid int64) error {
	err := l.delete(ctx, uid)
	if err != nil {
		if errors.Is(err, errno.ErrUserNotFound) {
			return err
		}
		logx.Errors(ctx, "user", "delete_self_failed", logx.Fields{
			"stage":   "delete_self",
			"user_id": uid,
			"error":   err.Error(),
		})
		return err
	}
	logx.Infos(ctx, "user", "delete_self_success", logx.Fields{
		"stage":   "delete_self",
		"user_id": uid,
	})
	return nil
}

func (l *user) AdminDelete(ctx context.Context, adminID, targetUID int64) error {
	err := l.delete(ctx, targetUID)
	if err != nil {
		if errors.Is(err, errno.ErrUserNotFound) {
			return err
		}

		logx.Errors(ctx, "admin", "admin_delete_failed", logx.Fields{
			"stage":      "delete_admin",
			"admin_id":   adminID,
			"target_uid": targetUID,
			"error":      err,
		})
		return err
	}
	logx.Infos(ctx, "admin", "admin_delete_success", logx.Fields{
		"stage":      "delete_admin",
		"admin_id":   adminID,
		"target_uid": targetUID,
	})
	return nil
}

func (l *user) delete(ctx context.Context, uid int64) error {
	userEntity, err := l.usersModel.FindOne(ctx, uid)
	if err != nil {
		// DB 查询失败 → 系统异常
		return fmt.Errorf("find user failed: %w", err)
	}

	if userEntity == nil {
		return errno.ErrUserNotFound
	}

	if err := l.usersModel.Delete(ctx, uid); err != nil {
		// 删除失败 → 系统异常
		return fmt.Errorf("delete user failed: %w", err)
	}

	return nil
}

func (l *user) List(ctx context.Context, req *domain.UserListReq) (*domain.UserListResp, error) {
	users, total, err := l.usersModel.List(ctx, req)
	if err != nil {
		logx.Errors(ctx, "user", "admin_get_list_failed", logx.Fields{
			"stage": "admin_get_list",
			"list":  req.Ids,
			"error": err.Error(),
		})
		return nil, err
	}

	resp := &domain.UserListResp{
		Count: total,
		List:  make([]*domain.User, 0, len(users)),
	}

	for _, u := range users {
		resp.List = append(resp.List, u.ToDomainUser())
	}

	return resp, nil
}
