package card

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"os"

	"github.com/skiplee85/card/dao"
	"github.com/skiplee85/card/msg"
	"github.com/skiplee85/common/route"
)

func List(req msg.ListCardReq) ([]*dao.Card, *route.Pagination) {
	data := []*dao.Card{}
	if req.Pagination == nil {
		req.Pagination = route.GetDefaultPagination()
	}
	req.Pagination.Total = 0
	query := bson.M{}
	dao.MgoExecCard(func(sc *mgo.Collection) {
		total, err := sc.Find(query).Count()
		if err == nil && total > 0 {
			sc.Find(query).Sort("no").Skip(req.Pagination.Size * (req.Pagination.Page - 1)).Limit(req.Pagination.Size).All(&data)
		}
		req.Pagination.Total = total
	})
	return data, req.Pagination
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
		f := fmt.Sprintf("data/pic/%s.jpg", req.NewNo)
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
