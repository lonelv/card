package handler

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
	"github.com/skiplee85/card/tesseract"
	"github.com/skiplee85/card/wx"
	"github.com/skiplee85/common/log"
	"github.com/skiplee85/common/route"
)

// RouteConf 路由配置
var RouteConf = []*route.BaseRoute{
	{
		Method:  "POST",
		Path:    "/upload",
		Handler: upload,
	},
	{
		Method:  "POST",
		Path:    "/send-img",
		Handler: sendImg,
	},
	{
		Method:  "POST",
		Path:    "/save-card",
		Handler: saveCard,
	},
	{
		Path: "/user",
		Child: []*route.BaseRoute{
			{
				// 登录
				Method:  "POST",
				Path:    "/login",
				Handler: login,
			},
		},
	},
	{
		Path: "/card",
		Role: 10,
		Child: []*route.BaseRoute{
			{
				Method:  "POST",
				Path:    "/list",
				Handler: listCard,
			},
			{
				Method:  "POST",
				Path:    "/modify",
				Handler: modifyCard,
			},
		},
	},
}

func upload(c *route.Context) {
	header, err := c.FormFile("pic")
	if err != nil {
		log.Error("%+v", err)
		return
	}
	file, err := header.Open()
	defer file.Close()

	imgByte, _ := ioutil.ReadAll(file)
	card := tesseract.ParseCardByBytes(imgByte)
	if card != nil {
		c.Send(card)
	} else {
		c.SendError(msg.ERROR_AUTH_CREDENTIALS_MISMATCH)
	}

}

func sendImg(c *route.Context) {
	var req msg.SendImgReq
	if err := c.ValidaArgs(&req); err != nil {
		return
	}
	if req.Data != "" {
		c.Finish(wx.SendNoticeImgBase64(req.OpenID, req.Data))
	} else if req.URL != "" {
		c.Finish(wx.SendNoticeImgURL(req.OpenID, req.URL))
	} else {
		c.SendError(msg.ERROR_REQUEST)
	}
}

func saveCard(c *route.Context) {
	var req msg.SaveCardReq
	if err := c.ValidaArgs(&req); err != nil {
		return
	}

	bs, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		log.Error("Not base64 img. Error: %v", err)
		c.SendError(msg.ERROR_REQUEST)
		return
	}
	f := fmt.Sprintf("data/pic/%s.jpg", req.No)

	file, err := os.OpenFile(f, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Error("%+v", err)
		c.SendError(msg.ERROR_INTERNAL)
		return
	}

	file.Write(bs)
	file.Close()

	card := &dao.Card{
		No:     req.No,
		Secret: req.Secret,
		Pic:    f,
		Create: time.Now(),
	}
	dao.MgoExecCard(func(sc *mgo.Collection) {
		_, err = sc.Upsert(bson.M{"no": card.No}, card)
	})
	if err != nil {
		log.Error("Mongo Error.%v", err)
		c.SendError(msg.ERROR_INTERNAL)
	} else {
		c.Send(card)
	}
}
