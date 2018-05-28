package dao

import (
	"time"
)

type User struct {
	UserID         int       `json:"user_id" bson:"_id"`
	Username       string    `json:"username" bson:"username"`
	Password       string    `json:"-" bson:"password"`
	Token          string    `json:"token" bson:"token"`
	Nickname       string    `json:"nickname" bson:"nickname"`
	Status         int       `json:"status" bson:"status"`
	Role           int       `json:"role" bson:"role"`
	CreateIP       string    `json:"create_ip" bson:"create_ip"`
	CreateTime     time.Time `json:"create_time" bson:"create_time"`
	LastLoginTime  time.Time `json:"last_login_time" bson:"last_login_time"` // 上次登陆时间
	LastLogOutTime time.Time `json:"last_logout_time" bson:"last_logout_time"`
	LoginTime      time.Time `json:"login_time" bson:"login_time"`
	LoginIP        string    `json:"login_ip" bson:"login_ip"`
}

// Card 卡密
type Card struct {
	No     int64     `json:"no" bson:"no"`
	Secret string    `json:"secret" bson:"secret"`
	Pic    string    `json:"pic" bson:"pic"`
	Create time.Time `json:"create" bson:"create"`
	Data   string    `json:"data" bson:"data"`
}
