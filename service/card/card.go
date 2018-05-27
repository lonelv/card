package card

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/skiplee85/card/dao"
	"github.com/skiplee85/card/msg"
	"github.com/skiplee85/common/log"
	"github.com/skiplee85/common/route"
)

const savePathFmt = "data/pic/%s.jpg"

func List(req msg.ListCardReq) ([]*dao.Card, *route.Pagination) {
	data := []*dao.Card{}
	if req.Pagination == nil {
		req.Pagination = route.GetDefaultPagination()
	}
	req.Pagination.Total = 0
	query := bson.M{}
	if req.No != "" {
		query["no"] = req.No
	}
	dao.MgoExecCard(func(sc *mgo.Collection) {
		total, err := sc.Find(query).Count()
		if err == nil && total > 0 {
			sc.Find(query).Sort("create").Skip(req.Pagination.Size * (req.Pagination.Page - 1)).Limit(req.Pagination.Size).All(&data)
		}
		req.Pagination.Total = total
	})
	return data, req.Pagination
}

func GetData(no string) (string, int) {
	var err error
	c := &dao.Card{}
	dao.MgoExecCard(func(sc *mgo.Collection) {
		err = sc.Find(bson.M{"no": no}).One(c)
	})
	if err != nil {
		log.Error("Get Card Error.%v", err)
		return "", msg.ERROR_REQUEST
	}
	if c.Data == "" {
		bs, err := ioutil.ReadFile(c.Pic)
		if err != nil {
			log.Error("Get Card Error.%v", err)
			return "", msg.ERROR_INTERNAL
		}
		c.Data = base64.StdEncoding.EncodeToString(bs)
		dao.MgoExecCard(func(sc *mgo.Collection) {
			err = sc.Update(bson.M{"no": no}, bson.M{"$set": bson.M{"data": c.Data}})
		})
	}
	return c.Data, msg.RET_OK
}

func Modify(req msg.ModifyCardReq) (*dao.Card, int) {
	var err error
	c := &dao.Card{}
	update := bson.M{}
	dao.MgoExecCard(func(sc *mgo.Collection) {
		err = sc.Find(bson.M{"no": req.No}).One(c)
	})
	if err != nil {
		return nil, msg.ERROR_REQUEST
	}
	if req.NewNo != "" && req.NewNo != req.No {
		f := fmt.Sprintf(savePathFmt, req.NewNo)
		os.Rename(c.Pic, f)
		update["no"] = req.NewNo
		update["pic"] = f
		c.No = req.NewNo
		c.Pic = f
	}
	if req.Secret != "" {
		update["secret"] = req.Secret
		c.Secret = req.Secret
	}
	dao.MgoExecCard(func(sc *mgo.Collection) {
		sc.Update(bson.M{"no": req.No}, bson.M{"$set": update})
	})
	return c, msg.RET_OK
}

func Save(req msg.SaveCardReq) (*dao.Card, int) {
	bs, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		log.Error("Not base64 img. Error: %v", err)
		return nil, msg.ERROR_REQUEST
	}
	f := fmt.Sprintf(savePathFmt, req.No)

	file, err := os.OpenFile(f, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Error("%+v", err)
		return nil, msg.ERROR_INTERNAL
	}

	file.Write(bs)
	file.Close()

	card := &dao.Card{
		No:     req.No,
		Secret: req.Secret,
		Pic:    f,
		Data:   req.Data,
		Create: time.Now(),
	}
	dao.MgoExecCard(func(sc *mgo.Collection) {
		_, err = sc.Upsert(bson.M{"no": card.No}, card)
	})
	if err != nil {
		log.Error("Mongo Error.%v", err)
		return nil, msg.ERROR_INTERNAL
	}
	return card, msg.RET_OK
}
