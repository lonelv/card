package dao

import (
	"time"
)

// Card 卡密
type Card struct {
	No     string    `bson:"no"`
	Secret string    `bson:"secret"`
	Pic    string    `bson:"pic"`
	Create time.Time `bson:"create"`
}
