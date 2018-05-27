package dao

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"

	"github.com/skiplee85/card/conf"
	"github.com/skiplee85/common/log"
	"github.com/skiplee85/common/mongodb"
)

var (
	MongoDB *mongodb.DialContext
	dbName  string
)

func InitMongo() {
	url := fmt.Sprintf("mongodb://%s:%s@%s:%d", conf.DB.User, conf.DB.Password, conf.DB.Host, conf.DB.Port)
	var err error
	MongoDB, err = mongodb.Dial(url, conf.DB.Sessions)
	if err != nil {
		log.Fatal("dial mongodb error: %v. URL: %s", err, url)
		panic(1)
	}
	log.Release("Connected to mongodb by " + url)
	dbName = conf.DB.Database

	MongoDB.EnsureUniqueIndex(dbName, "user", []string{"username"})
	MongoDB.EnsureCounter(dbName, "counter", "user")
	MongoDB.EnsureUniqueIndex(dbName, "card", []string{"no"})
	checkUser()
}

func MgoExec(collection string, f func(sc *mgo.Collection)) {
	if MongoDB == nil {
		log.Error("MongoDB is not init yet")
		return
	}

	s := MongoDB.Ref()
	defer MongoDB.UnRef(s)

	session := s.DB(dbName).C(collection)
	f(session)
}

func MgoExecUser(f func(sc *mgo.Collection)) {
	MgoExec("user", f)
}

func MgoExecCard(f func(sc *mgo.Collection)) {
	MgoExec("card", f)
}

func checkUser() {
	var u User
	MgoExecUser(func(sc *mgo.Collection) {
		err := sc.Find(bson.M{"username": "admin"}).One(&u)
		if err == mgo.ErrNotFound {
			userID, err := MongoDB.NextSeq(dbName, "counter", "user")
			if err != nil {
				return
			}

			hash, err := bcrypt.GenerateFromPassword([]byte("admin123321"), bcrypt.DefaultCost)
			if err != nil {
				return
			}
			u.Username = "admin"
			u.UserID = userID
			u.Password = string(hash)
			u.Role = 999
			t := time.Now()
			u.CreateTime = t
			u.LastLoginTime = t
			u.LastLogOutTime = t
			u.LoginTime = t
			sc.Insert(u)
		}
	})
}
