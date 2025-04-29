package crypto

import (
    "crypto/md5"
    "encoding/hex"
    "errors"
)

// ErrPasswordMismatch 密码不匹配错误
var ErrPasswordMismatch = errors.New("密码不匹配")

// HashPassword 生成MD5密码哈希
// 参数: password 明文密码
// 返回: 32位小写MD5哈希字符串
func HashPassword(password string) string {
    hash := md5.New()
    hash.Write([]byte(password))
    return hex.EncodeToString(hash.Sum(nil))
}

// VerifyPassword 验证明文密码与哈希值是否一致
// 参数:
//   - password: 明文密码
//   - hash: 数据库中保存的哈希值
// 返回: 验证通过返回nil，否则返回ErrPasswordMismatch
func VerifyPassword(password, hash string) error {
    hashed := HashPassword(password)
    if hashed != hash {
        return ErrPasswordMismatch
    }
    return nil
}
