// 自动生成的handler文件，请根据需要修改

package user

import (
	"context"
	"time"
	"crypto/md5"
	"fmt"
	"encoding/hex"

	"github.com/cloudwego/hertz/pkg/app"
	"novelai/pkg/constants"
	
	"novelai/biz/model/user"
	"novelai/biz/dal/db"
)

// 生成MD5密码哈希
func generatePasswordHash(password string) string {
	hash := md5.New()
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil))
}

// 生成简单的令牌 (实际生产环境应使用更安全的JWT或其他方式)
func generateToken(userID int64, username string) string {
	timestamp := time.Now().Unix()
	token := fmt.Sprintf("%d_%s_%d", userID, username, timestamp)
	hash := md5.New()
	hash.Write([]byte(token))
	return hex.EncodeToString(hash.Sum(nil))
}

// 用户注册
func Register(ctx context.Context, c *app.RequestContext) {
	// 获取请求参数
	req := new(user.RegisterRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(constants.StatusBadRequest, &user.RegisterResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// 校验参数
	if req.Username == "" || req.Password == "" {
		c.JSON(constants.StatusBadRequest, &user.RegisterResponse{
			Code:    400,
			Message: "用户名和密码不能为空",
		})
		return
	}

	// 创建用户对象
	newUser := &db.User{
		Username: req.Username,
		Password: generatePasswordHash(req.Password), // 密码加密存储
		Nickname: req.Nickname,
		Email:    req.Email,
		Status:   1, // 默认正常状态
	}

	// 调用数据库操作创建用户
	userID, err := db.CreateUser(newUser)
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

	// 生成token
	token := generateToken(userID, req.Username)
	
	// 返回成功响应
	c.JSON(constants.StatusOK, &user.RegisterResponse{
		Code:    0,
		Message: "注册成功",
		UserId:  userID,
		Token:   token,
	})
}

// 用户登录
func Login(ctx context.Context, c *app.RequestContext) {
	// 获取请求参数
	req := new(user.LoginRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(constants.StatusBadRequest, &user.LoginResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// 校验参数
	if req.Username == "" || req.Password == "" {
		c.JSON(constants.StatusBadRequest, &user.LoginResponse{
			Code:    400,
			Message: "用户名和密码不能为空",
		})
		return
	}

	// 验证用户名和密码
	hashPassword := generatePasswordHash(req.Password)
	userID, err := db.VerifyUser(req.Username, hashPassword)
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

	// 验证用户存在性（无需使用返回的userInfo）
	_, err = db.QueryUserByID(userID)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, &user.LoginResponse{
			Code:    500,
			Message: "获取用户信息失败",
		})
		return
	}

	// 生成token
	token := generateToken(userID, req.Username)
	
	// 返回成功响应
	c.JSON(constants.StatusOK, &user.LoginResponse{
		Code:    0,
		Message: "登录成功",
		UserId:  userID,
		Token:   token,
	})
}

// 获取用户信息
func GetUser(ctx context.Context, c *app.RequestContext) {
	// 获取请求参数
	req := new(user.GetUserRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(constants.StatusBadRequest, &user.GetUserResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// 验证用户ID
	if req.UserId <= 0 {
		c.JSON(constants.StatusBadRequest, &user.GetUserResponse{
			Code:    400,
			Message: "无效的用户ID",
		})
		return
	}

	// 获取用户信息
	userInfo, err := db.QueryUserByID(req.UserId)
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

	// 构建返回的用户对象
	userResp := &user.User{
		Id:       userInfo.ID,
		Username: userInfo.Username,
		Nickname: userInfo.Nickname,
		Avatar:   userInfo.Avatar,
		Email:    userInfo.Email,
		Status:   int32(userInfo.Status),
		CreatedAt: userInfo.CreatedAt.Unix(),
		UpdatedAt: userInfo.UpdatedAt.Unix(),
	}
	
	// 返回成功响应
	c.JSON(constants.StatusOK, &user.GetUserResponse{
		Code:    0,
		Message: "获取成功",
		User:    userResp,
	})
}

// 更新用户信息
func UpdateUser(ctx context.Context, c *app.RequestContext) {
	// 获取请求参数
	req := new(user.UpdateUserRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(constants.StatusBadRequest, &user.UpdateUserResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// 验证用户ID
	if req.UserId <= 0 {
		c.JSON(constants.StatusBadRequest, &user.UpdateUserResponse{
			Code:    400,
			Message: "无效的用户ID",
		})
		return
	}

	// 首先检查用户是否存在
	exists, err := db.CheckUserExists(req.UserId)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, &user.UpdateUserResponse{
			Code:    500,
			Message: "验证用户失败：" + err.Error(),
		})
		return
	}

	if !exists {
		c.JSON(constants.StatusOK, &user.UpdateUserResponse{
			Code:    1003,
			Message: "用户不存在",
		})
		return
	}

	// 创建更新对象
	updateUser := &db.User{
		ID:       req.UserId,
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		Email:    req.Email,
	}

	// 更新用户信息
	err = db.UpdateUserProfile(updateUser)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, &user.UpdateUserResponse{
			Code:    500,
			Message: "更新用户信息失败：" + err.Error(),
		})
		return
	}
	
	// 返回成功响应
	c.JSON(constants.StatusOK, &user.UpdateUserResponse{
		Code:    0,
		Message: "更新成功",
	})
}
