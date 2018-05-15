package wx

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"gopkg.in/chanxuehong/wechat.v2/mp/core"
	"gopkg.in/chanxuehong/wechat.v2/mp/menu"
	"gopkg.in/chanxuehong/wechat.v2/mp/message/callback/request"
	"gopkg.in/chanxuehong/wechat.v2/mp/message/callback/response"
	"gopkg.in/chanxuehong/wechat.v2/mp/message/custom"

	"github.com/skiplee85/card/dao"
	"github.com/skiplee85/card/log"
	"github.com/skiplee85/card/tesseract"
)

type WXConfig struct {
	WXAppID        string `ini:"WX_APP_ID"`
	WXAppSecret    string `ini:"WX_APP_SECRET"`
	WXToken        string `ini:"WX_TOKEN"`
	WXOriID        string `ini:"WX_ORI_ID"`
	WXEncodeAESKey string `ini:"WX_ENCODE_AES_KEY"`
}

var (
	msgHandler        core.Handler
	msgServer         *core.Server
	accessTokenServer core.AccessTokenServer
	wechatClient      *core.Client
)

// InitWX 初始化微信
func InitWX(config *WXConfig) {
	mux := core.NewServeMux()
	mux.DefaultMsgHandleFunc(defaultMsgHandler)
	mux.DefaultEventHandleFunc(defaultEventHandler)
	mux.MsgHandleFunc(request.MsgTypeText, textMsgHandler)
	mux.MsgHandleFunc(request.MsgTypeImage, imgMsgHandler)
	mux.EventHandleFunc(menu.EventTypeClick, menuClickEventHandler)

	msgHandler = mux
	msgServer = core.NewServer(config.WXOriID, config.WXAppID, config.WXToken, config.WXEncodeAESKey, msgHandler, nil)

	accessTokenServer = core.NewDefaultAccessTokenServer(config.WXAppID, config.WXAppSecret, nil)
	wechatClient = core.NewClient(accessTokenServer, nil)

	http.HandleFunc("/wx_callback", wxCallbackHandler)
}

func textMsgHandler(ctx *core.Context) {
	log.Debug("收到文本消息:\n%s\n", ctx.MsgPlaintext)

	msg := request.GetText(ctx.MixedMsg)
	resp := response.NewText(msg.FromUserName, msg.ToUserName, msg.CreateTime, msg.Content)
	ctx.RawResponse(resp) // 明文回复
	// ctx.AESResponse(resp, 0, "", nil) // aes密文回复
}

func imgMsgHandler(ctx *core.Context) {
	log.Debug("收到图片消息:\n%s\n", ctx.MsgPlaintext)
	ctx.NoneResponse()
	go func() {
		msg := request.GetImage(ctx.MixedMsg)
		ret := ""
		c := getPic(msg.PicURL)
		if c == nil {
			ret = "解析失败，请重拍~"
		} else {
			bs, _ := json.Marshal(c)
			ret = string(bs)
		}
		resp := custom.NewText(msg.ToUserName, ret, "")
		custom.Send(wechatClient, resp)
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

func defaultEventHandler(ctx *core.Context) {
	log.Debug("收到事件:\n%s\n", ctx.MsgPlaintext)
	ctx.NoneResponse()
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
