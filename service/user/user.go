package user

import (
	"golang.org/x/crypto/bcrypt"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/skiplee85/card/conf"
	"github.com/skiplee85/card/dao"
	"github.com/skiplee85/card/msg"
	"github.com/skiplee85/common/log"
	"github.com/skiplee85/common/route"
)

func Login(username, password, ip string) (*dao.User, int) {
	var user dao.User
	var err error
	dao.MgoExecUser(func(sc *mgo.Collection) {
		err = sc.Find(bson.M{"username": username}).One(&user)
	})
	if err != nil {
		if err != mgo.ErrNotFound {
			log.Error("func Login. %+v", err)
		}
		return nil, msg.ERROR_USER_NOT_FOUND
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, msg.ERROR_INVALID_PASSWORD
	}
	_, user.Token, err = route.GenJwtToken(user.UserID, user.Role, conf.TokenExpire)
	if err != nil {
		log.Error("call genToken error. %v", err)
		return nil, msg.ERROR_INTERNAL
	}

	user.LoginIP = ip
	user.LastLoginTime = user.LoginTime
	user.LoginTime = time.Now()
	dao.MgoExecUser(func(sc *mgo.Collection) {
		err = sc.Update(bson.M{"_id": user.UserID}, bson.M{"$set": bson.M{
			"token":           user.Token,
			"login_ip":        user.LoginIP,
			"login_time":      user.LoginTime,
			"last_login_time": user.LastLoginTime,
		}})
	})
	if err != nil {
		log.Error("call genToken error. %v", err)
		return nil, msg.ERROR_INTERNAL
	}

	return &user, msg.RET_OK
}
