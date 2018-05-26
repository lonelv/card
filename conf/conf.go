package conf

import (
	"github.com/go-ini/ini"
)

const (
	// TokenExpire Token过期时间
	TokenExpire = 30 * 24 * 60 * 60 //30 天
)

// Common 通用配置
var Common struct {
	Level     string
	JWTSecret string
}

// DB 数据库配置
var DB struct {
	User     string
	Password string
	Host     string
	Port     int
	DataBase string
	Sessions int
}

// HTTP 网络配置
var HTTP struct {
	Address string
	Port    int
}

// WX 微信配置
var WX struct {
	AppID        string
	AppSecret    string
	Token        string
	OriID        string
	EncodeAESKey string
}

func init() {
	cfg, err := ini.Load(".env")
	if err != nil {
		panic(err)
	}
	err = cfg.Section("Common").MapTo(&Common)
	if err != nil {
		panic(err)
	}
	err = cfg.Section("DB").MapTo(&DB)
	if err != nil {
		panic(err)
	}
	err = cfg.Section("Http").MapTo(&HTTP)
	if err != nil {
		panic(err)
	}
	err = cfg.Section("WX").MapTo(&WX)
	if err != nil {
		panic(err)
	}
}
