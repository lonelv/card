package msg

import (
	"github.com/skiplee85/common/route"
)

const (
	RET_OK                  = route.CodeOk
	ERROR_REQUEST           = route.CodeErrorRequest     //非法请求
	ERROR_INTERNAL          = route.CodeErrorInternal    //服务器内部错误
	ERROR_INVALID_ARGUMENTS = route.CodeInvalidArguments // 非法参数

	ERROR_AUTH_CREDENTIALS_MISMATCH = 4001001 // 验证失败

	// user
	ERROR_INVALID_PASSWORD     = 202001 // 错误的密码
	ERROR_INVALID_OLD_PASSWORD = 202002 // 原密码不匹配
	ERROR_USER_NOT_FOUND       = 203001 // 用户不存在
	ERROR_USER_NAME_EXISTS     = 203002 // 用户名已被注册
	ERROR_FORBIDDEN_USER       = 203003 // 账号已被封
)

// LoginReq 管理后台登陆
type LoginReq struct {
	Username string `json:"username" form:"username" binding:"required,min=4,max=16"`
	Password string `json:"password" form:"password" binding:"required,min=6,max=32"`
}

// PasswordReq 管理后台修改密码请求参数
type PasswordReq struct {
	OriPwd string `json:"password" form:"password" binding:"required,min=6,max=32"`
	NewPwd string `json:"new_password" form:"new_password" binding:"required,min=6,max=32"`
}

// SendImgReq 发送图片到微信
type SendImgReq struct {
	OpenID string `json:"openid" form:"openid" binding:"required"`
	Data   string `json:"data" form:"data"`
	URL    string `json:"url" form:"url"`
}

// SaveCardReq 发送图片到微信
type SaveCardReq struct {
	No     string `json:"no" form:"no"`
	Data   string `json:"data" form:"data" binding:"required"`
	Secret string `json:"secret" form:"secret"`
}

type ListCardReq struct {
	No         string            `json:"no" form:"no"`
	Pagination *route.Pagination `json:"pagination" form:"pagination"`
}

type ModifyCardReq struct {
	No     int64  `json:"no" form:"no" binding:"required"`
	NewNo  int64  `json:"new_no" form:"new_no"`
	Secret string `json:"secret" form:"secret"`
}
