package dao

import (
	"fmt"
	"gopkg.in/mgo.v2"

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
	dbName = conf.DB.DataBase

	MongoDB.EnsureUniqueIndex(dbName, "card", []string{"no"})
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

func MgoExecCard(f func(sc *mgo.Collection)) {
	MgoExec("card", f)
}

func OnDestroy() {
	if MongoDB != nil {
		log.Debug("Close mongoDB.") // TODO: mongoDB is nil. Why?
		MongoDB.Close()
		MongoDB = nil
	}
}
