// 自动生成的handler文件，请根据需要修改

package user

import (
	"context"
	"crypto/md5"
	"encoding/hex"

	"github.com/cloudwego/hertz/pkg/app"

	"novelai/pkg/constants"
	middleware "novelai/pkg/middleware"

	userpb "novelai/biz/model/user"
	"novelai/biz/dal/db"
	service "novelai/biz/service/user"
)

// FIXME: import order, ensure standard, third-party, local ordering

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
	req := new(userpb.RegisterRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(constants.StatusBadRequest, &userpb.RegisterResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}
	if req.Username == "" || req.Password == "" {
		c.JSON(constants.StatusBadRequest, &userpb.RegisterResponse{
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
			c.JSON(constants.StatusOK, &userpb.RegisterResponse{
				Code:    1001,
				Message: "用户名已存在",
			})
			return
		}
		c.JSON(constants.StatusInternalServerError, &userpb.RegisterResponse{
			Code:    500,
			Message: "注册失败：" + err.Error(),
		})
		return
	}
	// 3. 注册成功，完整返回所有字段
	c.JSON(constants.StatusOK, &userpb.RegisterResponse{
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
	req := new(userpb.LoginRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(constants.StatusBadRequest, &userpb.LoginResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}
	if req.Username == "" || req.Password == "" {
		c.JSON(constants.StatusBadRequest, &userpb.LoginResponse{
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
			c.JSON(constants.StatusOK, &userpb.LoginResponse{
				Code:    1002,
				Message: "用户名或密码错误",
			})
			return
		}
		c.JSON(constants.StatusInternalServerError, &userpb.LoginResponse{
			Code:    500,
			Message: "登录失败：" + err.Error(),
		})
		return
	}
	// 登录成功，完整返回所有字段，确保 user_id 字段为 int64 且 json tag 为 user_id
	resp := &userpb.LoginResponse{
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
	req := new(userpb.GetUserRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(constants.StatusBadRequest, &userpb.GetUserResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}
	// 统一从 JWT 获取 userId，避免前端传递
	idVal, _ := c.Get(middleware.IdentityKey)
	userId := idVal.(int64)
	svc := service.NewUserService(ctx, c)
	userResp, err := svc.GetUserInfo(userId)
	if err != nil {
		if err == db.ErrUserNotFound {
			c.JSON(constants.StatusOK, &userpb.GetUserResponse{
				Code:    1003,
				Message: "用户不存在",
			})
			return
		}
		c.JSON(constants.StatusInternalServerError, &userpb.GetUserResponse{
			Code:    500,
			Message: "获取用户信息失败：" + err.Error(),
		})
		return
	}
	c.JSON(constants.StatusOK, &userpb.GetUserResponse{
		Code:    200,
		Message: "获取成功",
		User:    userResp,
	})
}

// 更新用户信息
// 只做参数校验和调用service层，所有业务逻辑下沉到service
func UpdateUser(ctx context.Context, c *app.RequestContext) {
	req := new(userpb.UpdateUserRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(constants.StatusBadRequest, &userpb.UpdateUserResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}
	// 统一从 JWT 获取 userId，避免前端传递
	idVal, _ := c.Get(middleware.IdentityKey)
	userId := idVal.(int64)
	svc := service.NewUserService(ctx, c)
	err := svc.UpdateUserProfile(userId, req)
	if err != nil {
		if err == db.ErrUserNotFound {
			c.JSON(constants.StatusOK, &userpb.UpdateUserResponse{
				Code:    1003,
				Message: "用户不存在",
			})
			return
		}
		c.JSON(constants.StatusInternalServerError, &userpb.UpdateUserResponse{
			Code:    500,
			Message: "更新用户信息失败：" + err.Error(),
		})
		return
	}
	c.JSON(constants.StatusOK, &userpb.UpdateUserResponse{
		Code:    200,
		Message: "更新成功",
	})
}

// ChangePassword 修改用户密码
func ChangePassword(ctx context.Context, c *app.RequestContext) {
	// 请求体绑定
	type changePasswordReq struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	req := new(changePasswordReq)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(constants.StatusBadRequest, &userpb.UpdateUserResponse{Code: constants.StatusBadRequest, Message: err.Error()})
		return
	}
	if req.OldPassword == "" || req.NewPassword == "" {
		c.JSON(constants.StatusBadRequest, &userpb.UpdateUserResponse{Code: constants.StatusBadRequest, Message: "旧密码和新密码不能为空"})
		return
	}
	// 密码哈希
	oldHash := generatePasswordHash(req.OldPassword)
	newHash := generatePasswordHash(req.NewPassword)
	// 获取用户ID
	idVal, _ := c.Get(middleware.IdentityKey)
	userId := idVal.(int64)
	// 调用服务
	svc := service.NewUserService(ctx, c)
	err := svc.UpdateUserPassword(userId, oldHash, newHash)
	if err != nil {
		if err == db.ErrInvalidPassword {
			c.JSON(constants.StatusOK, &userpb.UpdateUserResponse{Code: 1002, Message: "旧密码错误"})
			return
		}
		c.JSON(constants.StatusInternalServerError, &userpb.UpdateUserResponse{Code: constants.StatusInternalServerError, Message: "密码修改失败：" + err.Error()})
		return
	}
	c.JSON(constants.StatusOK, &userpb.UpdateUserResponse{Code: constants.StatusOK, Message: "密码修改成功"})
}

// DeleteUser 删除当前用户（软删除）
func DeleteUser(ctx context.Context, c *app.RequestContext) {
	// 获取用户ID
	idVal, _ := c.Get(middleware.IdentityKey)
	userId := idVal.(int64)
	svc := service.NewUserService(ctx, c)
	err := svc.DeleteUser(userId)
	if err != nil {
		if err == db.ErrUserNotFound {
			c.JSON(constants.StatusOK, &userpb.UpdateUserResponse{Code: 1003, Message: "用户不存在"})
			return
		}
		c.JSON(constants.StatusInternalServerError, &userpb.UpdateUserResponse{Code: constants.StatusInternalServerError, Message: "删除用户失败：" + err.Error()})
		return
	}
	c.JSON(constants.StatusOK, &userpb.UpdateUserResponse{Code: constants.StatusOK, Message: "删除成功"})
}
