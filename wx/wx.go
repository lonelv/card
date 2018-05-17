package wx

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"gopkg.in/chanxuehong/wechat.v2/mp/core"
	"gopkg.in/chanxuehong/wechat.v2/mp/media"
	"gopkg.in/chanxuehong/wechat.v2/mp/menu"
	"gopkg.in/chanxuehong/wechat.v2/mp/message/callback/request"
	"gopkg.in/chanxuehong/wechat.v2/mp/message/callback/response"
	"gopkg.in/chanxuehong/wechat.v2/mp/message/custom"
	"gopkg.in/chanxuehong/wechat.v2/mp/message/template"

	"github.com/skiplee85/card/dao"
	"github.com/skiplee85/card/log"
	"github.com/skiplee85/card/tesseract"
)

var (
	msgHandler        core.Handler
	msgServer         *core.Server
	accessTokenServer core.AccessTokenServer
	wechatClient      *core.Client
)

type standard struct {
	Value string `json:"value"`
	Color string `json:"color"`
}

type tpl struct {
	First    standard `json:"first"`
	Keyword1 standard `json:"keyword1"`
	Keyword2 standard `json:"keyword2"`
	Remark   standard `json:"remark"`
}

// InitWXServer 初始化微信服务
func InitWXServer(wxAppID, wxToken, wxOriID, wxEncodeAESKey string) {
	mux := core.NewServeMux()
	mux.DefaultMsgHandleFunc(defaultMsgHandler)
	mux.DefaultEventHandleFunc(defaultEventHandler)
	mux.MsgHandleFunc(request.MsgTypeText, textMsgHandler)
	mux.MsgHandleFunc(request.MsgTypeImage, imgMsgHandler)
	mux.EventHandleFunc(menu.EventTypeClick, menuClickEventHandler)
	mux.EventHandleFunc(request.EventTypeSubscribe, subscribeEventHandler)

	msgHandler = mux
	msgServer = core.NewServer(wxOriID, wxAppID, wxToken, wxEncodeAESKey, msgHandler, nil)

	http.HandleFunc("/wx_callback", wxCallbackHandler)
}

// InitWXClient 初始化微信操作
func InitWXClient(wxAppID, wxAppSecret string) {
	accessTokenServer = core.NewDefaultAccessTokenServer(wxAppID, wxAppSecret, nil)
	wechatClient = core.NewClient(accessTokenServer, nil)
}

func textMsgHandler(ctx *core.Context) {
	log.Debug("收到文本消息:\n%s\n", ctx.MsgPlaintext)

	msg := request.GetText(ctx.MixedMsg)
	resp := response.NewText(msg.FromUserName, msg.ToUserName, msg.CreateTime, fmt.Sprintf("你的openid: %s", msg.FromUserName))
	ctx.RawResponse(resp) // 明文回复
	// ctx.AESResponse(resp, 0, "", nil) // aes密文回复
}

func imgMsgHandler(ctx *core.Context) {
	log.Debug("收到图片消息:\n%s\n", ctx.MsgPlaintext)
	ctx.NoneResponse()
	go func() {
		msg := request.GetImage(ctx.MixedMsg)
		data := &tpl{}
		data.Keyword1.Value = "系统"
		c := getPic(msg.PicURL)
		if c == nil {
			data.First.Value = "解析失败"
			data.Keyword2.Value = "请重新拍摄"
		} else {
			data.First.Value = "解析成功"
			data.Keyword2.Value = fmt.Sprintf("卡号:%s\n卡密:%s", c.No, c.Secret)
		}
		resp := &template.TemplateMessage2{
			ToUser:     msg.FromUserName,
			TemplateId: "9YhtUXt4qIs7h_qtcungbN0dGxwdgn5B4w8Nk-RDW9U",
			Data:       data,
		}
		template.Send(wechatClient, resp)
	}()
}

func defaultMsgHandler(ctx *core.Context) {
	log.Debug("收到消息:\n%s\n", ctx.MsgPlaintext)
	ctx.NoneResponse()
}

func menuClickEventHandler(ctx *core.Context) {
	log.Debug("收到菜单 click 事件:\n%s\n", ctx.MsgPlaintext)

	event := menu.GetClickEvent(ctx.MixedMsg)
	resp := response.NewText(event.FromUserName, event.ToUserName, event.CreateTime, "收到 click 类型的事件")
	ctx.RawResponse(resp) // 明文回复
	// ctx.AESResponse(resp, 0, "", nil) // aes密文回复
}

func subscribeEventHandler(ctx *core.Context) {
	log.Debug("收到关注事件:\n%s\n", ctx.MsgPlaintext)

	event := menu.GetClickEvent(ctx.MixedMsg)
	resp := response.NewText(event.FromUserName, event.ToUserName, event.CreateTime, fmt.Sprintf("你的openid: %s", event.FromUserName))
	ctx.RawResponse(resp) // 明文回复
	// ctx.AESResponse(resp, 0, "", nil) // aes密文回复
}

func defaultEventHandler(ctx *core.Context) {
	log.Debug("收到事件:\n%s\n", ctx.MsgPlaintext)
	ctx.NoneResponse()
}

// SendNoticeImgURL 发送图片
func SendNoticeImgURL(openid, url string) {
	imgResp, err := http.Get(url)
	if err != nil {
		log.Error("Get img Error:%+v", err)
		return
	}
	defer imgResp.Body.Close()

	info, err := media.UploadImageFromReader(wechatClient, fmt.Sprintf("%d.jpg", time.Now().UnixNano()), imgResp.Body)
	if err != nil {
		log.Error("Upload Error:%+v", err)
		return
	}
	err = custom.Send(wechatClient, custom.NewImage(openid, info.MediaId, ""))
	if err != nil {
		log.Error("SendNotice Error:%+v", err)
	}
}

// SendNoticeImgBase64 发送图片
func SendNoticeImgBase64(openid, data string) {
	// 先发个模板消息通知下
	t := &tpl{}
	t.First.Value = "抢到单啦！"
	t.First.Color = "#173177"
	t.Keyword1.Value = "系统"
	t.Keyword2.Value = "还不赶紧去支付~"

	resp := &template.TemplateMessage2{
		ToUser:     openid,
		TemplateId: "9YhtUXt4qIs7h_qtcungbN0dGxwdgn5B4w8Nk-RDW9U",
		Data:       t,
	}
	template.Send(wechatClient, resp)

	bs, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		log.Error("Not base64 img. Error: %d", err)
		return
	}
	tmpFile := fmt.Sprintf("%d.jpg", time.Now().UnixNano())
	f, err := os.OpenFile(tmpFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Error("%+v", err)
		f.Close()
		os.Remove(tmpFile)
		return
	}

	f.Write(bs)
	f.Close()

	info, err := media.UploadImage(wechatClient, tmpFile)
	if err != nil {
		log.Error("Upload Error:%+v", err)
		return
	}
	err = custom.Send(wechatClient, custom.NewImage(openid, info.MediaId, ""))
	if err != nil {
		log.Error("SendNotice Error:%+v", err)
	}
	os.Remove(tmpFile)
}

// wxCallbackHandler 是处理回调请求的 http handler.
//  1. 不同的 web 框架有不同的实现
//  2. 一般一个 handler 处理一个公众号的回调请求(当然也可以处理多个, 这里我只处理一个)
func wxCallbackHandler(w http.ResponseWriter, r *http.Request) {
	msgServer.ServeHTTP(w, r, nil)
}

func getPic(url string) *dao.Card {
	imgResp, err := http.Get(url)
	if err != nil {
		log.Error("Get img Error:%+v", err)
		return nil
	}
	defer imgResp.Body.Close()

	imgByte, _ := ioutil.ReadAll(imgResp.Body)
	return tesseract.ParseCardByBytes(imgByte)
}
