// 自动生成的handler文件，请根据需要修改

package user

import (
	"context"
	"crypto/md5"
	"encoding/hex"

	"github.com/cloudwego/hertz/pkg/app"
	"novelai/pkg/constants"
	
	"novelai/biz/model/user"
	"novelai/biz/dal/db"
	service "novelai/biz/service/user"
)

// 生成MD5密码哈希
func generatePasswordHash(password string) string {
	hash := md5.New()
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil))
}

// 已废弃：原generateToken函数
// 说明：令牌生成与校验已由hertz-contrib/jwt中间件统一处理，业务代码无需手写token逻辑。
// 在路由注册阶段配置jwt中间件，登录接口自动生成JWT，受保护接口自动校验。
// 详见路由注册与中间件配置。

// 用户注册
// 只做参数校验和调用service层，所有业务逻辑下沉到service
func Register(ctx context.Context, c *app.RequestContext) {
	// 1. 参数校验
	req := new(user.RegisterRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(constants.StatusBadRequest, &user.RegisterResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}
	if req.Username == "" || req.Password == "" {
		c.JSON(constants.StatusBadRequest, &user.RegisterResponse{
			Code:    400,
			Message: "用户名和密码不能为空",
		})
		return
	}

	// 3. 调用service
	svc := service.NewUserService(ctx, c)
	userID, _, err := svc.Register(req)
	if err != nil {
		if err == db.ErrUserAlreadyExists {
			c.JSON(constants.StatusOK, &user.RegisterResponse{
				Code:    1001,
				Message: "用户名已存在",
			})
			return
		}
		c.JSON(constants.StatusInternalServerError, &user.RegisterResponse{
			Code:    500,
			Message: "注册失败：" + err.Error(),
		})
		return
	}
	c.JSON(constants.StatusOK, &user.RegisterResponse{
		Code:    0,
		Message: "注册成功",
		UserId:  userID,
	})
}

// 用户登录
// 只做参数校验和调用service层，所有业务逻辑下沉到service
func Login(ctx context.Context, c *app.RequestContext) {
	req := new(user.LoginRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(constants.StatusBadRequest, &user.LoginResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}
	if req.Username == "" || req.Password == "" {
		c.JSON(constants.StatusBadRequest, &user.LoginResponse{
			Code:    400,
			Message: "用户名和密码不能为空",
		})
		return
	}
	req.Password = generatePasswordHash(req.Password)
	svc := service.NewUserService(ctx, c)
	userID, _, err := svc.Login(req)
	if err != nil {
		if err == db.ErrInvalidPassword || err == db.ErrUserNotFound {
			c.JSON(constants.StatusOK, &user.LoginResponse{
				Code:    1002,
				Message: "用户名或密码错误",
			})
			return
		}
		c.JSON(constants.StatusInternalServerError, &user.LoginResponse{
			Code:    500,
			Message: "登录失败：" + err.Error(),
		})
		return
	}
	c.JSON(constants.StatusOK, &user.LoginResponse{
		Code:    0,
		Message: "登录成功",
		UserId:  userID,
	})
}

// 获取用户信息
// 只做参数校验和调用service层，所有业务逻辑下沉到service
func GetUser(ctx context.Context, c *app.RequestContext) {
	req := new(user.GetUserRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(constants.StatusBadRequest, &user.GetUserResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}
	if req.UserId <= 0 {
		c.JSON(constants.StatusBadRequest, &user.GetUserResponse{
			Code:    400,
			Message: "无效的用户ID",
		})
		return
	}
	svc := service.NewUserService(ctx, c)
	userResp, err := svc.GetUserInfo(req.UserId)
	if err != nil {
		if err == db.ErrUserNotFound {
			c.JSON(constants.StatusOK, &user.GetUserResponse{
				Code:    1003,
				Message: "用户不存在",
			})
			return
		}
		c.JSON(constants.StatusInternalServerError, &user.GetUserResponse{
			Code:    500,
			Message: "获取用户信息失败：" + err.Error(),
		})
		return
	}
	c.JSON(constants.StatusOK, &user.GetUserResponse{
		Code:    0,
		Message: "获取成功",
		User:    userResp,
	})
}

// 更新用户信息
// 只做参数校验和调用service层，所有业务逻辑下沉到service
func UpdateUser(ctx context.Context, c *app.RequestContext) {
	req := new(user.UpdateUserRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(constants.StatusBadRequest, &user.UpdateUserResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}
	if req.UserId <= 0 {
		c.JSON(constants.StatusBadRequest, &user.UpdateUserResponse{
			Code:    400,
			Message: "无效的用户ID",
		})
		return
	}
	svc := service.NewUserService(ctx, c)
	err := svc.UpdateUserProfile(req.UserId, req)
	if err != nil {
		if err == db.ErrUserNotFound {
			c.JSON(constants.StatusOK, &user.UpdateUserResponse{
				Code:    1003,
				Message: "用户不存在",
			})
			return
		}
		c.JSON(constants.StatusInternalServerError, &user.UpdateUserResponse{
			Code:    500,
			Message: "更新用户信息失败：" + err.Error(),
		})
		return
	}
	c.JSON(constants.StatusOK, &user.UpdateUserResponse{
		Code:    0,
		Message: "更新成功",
	})
}
