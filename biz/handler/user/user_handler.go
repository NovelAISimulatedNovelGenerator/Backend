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
// 注册成功时响应完整 RegisterResponse，包含 code、message、user_id、token 字段，便于前端/自动化测试获取 token
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

	// 2. 调用 service 层注册逻辑，获取 userId 和 token
	svc := service.NewUserService(ctx, c)
	userID, token, err := svc.Register(req)
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
	// 3. 注册成功，完整返回所有字段
	c.JSON(constants.StatusOK, &user.RegisterResponse{
		Code:    200,
		Message: "注册成功",
		UserId:  userID,
		Token:   token,
	})
}


// 用户登录
// 只做参数校验和调用service层，所有业务逻辑下沉到service
// 登录成功时响应完整 LoginResponse，包含 code、message、user_id、token 字段，便于前端/自动化测试获取 token
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
	// 密码加密，与注册保持一致
	req.Password = generatePasswordHash(req.Password)
	svc := service.NewUserService(ctx, c)
	userID, token, err := svc.Login(req)
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
	// 登录成功，完整返回所有字段，确保 user_id 字段为 int64 且 json tag 为 user_id
	resp := &user.LoginResponse{
		Code:    200,
		Message: "登录成功",
		UserId:  userID, // int64 类型，json:"user_id"，与 proto 定义一致
		Token:   token,
	}
	// 关键注释：user.LoginResponse 的 json tag 必须为 user_id，且类型为 int64
	// 若前端/测试脚本仍无法获取 user_id，请检查 user.LoginResponse 结构体定义及 proto 文件
	c.JSON(constants.StatusOK, resp)
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
		Code:    200,
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
		Code:    200,
		Message: "更新成功",
	})
}
