package route

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
)

var routeConf = map[string][]baseRoute{
	"/": []baseRoute{
		{
			Method:  "POST",
			Path:    "/upload",
			Handler: upload,
		},
		{
			Method:  "POST",
			Path:    "/send-imgbase64",
			Handler: sendImg,
		},
		{
			Method:  "POST",
			Path:    "/save-card",
			Handler: saveCard,
		},
	},
	"/admin": []baseRoute{
		{
			// 登录
			Method:  "POST",
			Path:    "/login",
			Handler: nil,
		},
	},
}

func upload(c *Context) {
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

func sendImg(c *Context) {
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

func saveCard(c *Context) {
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
