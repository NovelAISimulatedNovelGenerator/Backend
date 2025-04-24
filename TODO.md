# TODO: 用户模块 JWT 一致性改造

- 切换登录路由 `/api/user/login` 为 `jwtMw.LoginHandler`，由中间件统一处理登录与 token 生成，删除手写 MD5 token 逻辑。
- 在 `middleware.JwtMiddleware` 中：
  - Authenticator 调用 `db.VerifyUser` 验证用户名与 MD5(password)，返回真实 `user_id`。
  - 自定义 `LoginResponse` 输出 `{code, message, user_id, token}`。
  - 自定义 `RefreshResponse` 输出 `{code, message, token}`。
- 更新受保护接口：
  - 从 `c.Get(middleware.IdentityKey)` 获取 `user_id`，移除请求体中的 `UserId` 字段。
- 补充其他路由：`/refresh`、`/logout`、`/change_password`、`/delete` 等。
