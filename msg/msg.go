package msg

import (
	jwt "github.com/dgrijalva/jwt-go"
)

const (
	RET_OK         = 0
	ERROR_REQUEST  = 400 //非法请求
	ERROR_INTERNAL = 500 //服务器内部错误

	// general: 1000 -2999
	ERROR_INVALID_ARGUMENTS = 1000 // 非法参数

	ERROR_AUTH_CREDENTIALS_MISMATCH = 4001001 // 验证失败

	// user
	ERROR_INVALID_PASSWORD     = 202001 // 错误的密码
	ERROR_INVALID_OLD_PASSWORD = 202002 // 原密码不匹配
	ERROR_USER_NOT_FOUND       = 203001 // 用户不存在
	ERROR_USER_NAME_EXISTS     = 203002 // 用户名已被注册
	ERROR_FORBIDDEN_USER       = 203003 // 账号已被封
)

type UserClaims struct {
	UserID   int
	Username string
	Role     int
	jwt.StandardClaims
}

// API 公共响应参数
type BaseResponse struct {
	Code       int         `json:"code"`
	Msg        string      `json:"msg,omitempty"`
	Data       interface{} `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Page  int `form:"page" json:"page" binding:"required,min=0"` // 第几页
	Size  int `form:"size" json:"size" binding:"required,min=0"` // 一页容纳最多容纳多少数据
	Total int `form:"total" json:"total,omitempty"`              // 共有多少数据
}

func GetDefaultPagination() *Pagination {
	return &Pagination{
		Page: 1,
		Size: 20,
	}
}

// LoginReq 管理后台登陆
type LoginReq struct {
	Username string `form:"username" json:"username" binding:"required,min=4,max=16"`
	Password string `form:"password" json:"password" binding:"required,min=6,max=32"`
}

// PasswordReq 管理后台修改密码请求参数
type PasswordReq struct {
	OriPwd string `json:"password" binding:"required,min=6,max=32"`
	NewPwd string `json:"new_password" binding:"required,min=6,max=32"`
}

// SendImgReq 发送图片到微信
type SendImgReq struct {
	OpenID string `json:"openid"`
	Data   string `json:"data"`
	URL    string `json:"url"`
}

// SaveCardReq 发送图片到微信
type SaveCardReq struct {
	No     string `json:"no"`
	Data   string `json:"data"`
	Secret string `json:"secret"`
}
