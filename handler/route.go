package handler

import (
	"io/ioutil"

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
			{
				Method:  "GET",
				Path:    "/get-data",
				Handler: getCardData,
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
