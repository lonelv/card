package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/chanxuehong/wechat.v2/mp/core"
	"gopkg.in/chanxuehong/wechat.v2/mp/menu"
	"gopkg.in/chanxuehong/wechat.v2/mp/message/callback/request"
	"gopkg.in/chanxuehong/wechat.v2/mp/message/callback/response"

	"github.com/go-ini/ini"
	"github.com/skiplee85/card/dao"
	"github.com/skiplee85/card/log"
	"github.com/skiplee85/card/tesseract"
)

type Config struct {
	Level          string `ini:"LEVEL"`
	DBUser         string `ini:"DB_USER"`
	DBPassword     string `ini:"DB_PASSWORD"`
	DBHost         string `ini:"DB_HOST"`
	DBPort         int    `ini:"DB_PORT"`
	DBDataBase     string `ini:"DB_DATABASE"`
	HTTPPort       int    `ini:"HTTP_PORT"`
	WXAppID        string `ini:"WX_APP_ID"`
	WXAppSecret    string `ini:"WX_APP_SECRET"`
	WXToken        string `ini:"WX_TOKEN"`
	WXOriID        string `ini:"WX_ORI_ID"`
	WXEncodeAESKey string `ini:"WX_ENCODE_AES_KEY"`
}

var (
	// 下面两个变量不一定非要作为全局变量, 根据自己的场景来选择.
	msgHandler core.Handler
	msgServer  *core.Server
)

func textMsgHandler(ctx *core.Context) {
	log.Debug("收到文本消息:\n%s\n", ctx.MsgPlaintext)

	msg := request.GetText(ctx.MixedMsg)
	resp := response.NewText(msg.FromUserName, msg.ToUserName, msg.CreateTime, msg.Content)
	ctx.RawResponse(resp) // 明文回复
	// ctx.AESResponse(resp, 0, "", nil) // aes密文回复
}

func imgMsgHandler(ctx *core.Context) {
	log.Debug("收到图片消息:\n%s\n", ctx.MsgPlaintext)

	msg := request.GetImage(ctx.MixedMsg)
	resp := response.NewText(msg.FromUserName, msg.ToUserName, msg.CreateTime, msg.PicURL)
	ctx.RawResponse(resp) // 明文回复
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

func main() {
	var config Config
	conf, err := ini.Load(".env")
	if err != nil {
		panic(err)
	}
	conf.BlockMode = false
	err = conf.MapTo(&config)
	if err != nil {
		panic(err)
	}

	log.InitLog(config.Level)
	dao.InitMongo(fmt.Sprintf("mongodb://%s:%s@%s:%d", config.DBUser, config.DBPassword, config.DBHost, config.DBPort), config.DBDataBase, 10)

	// wxAppSecret = config.WXAppSecret

	mux := core.NewServeMux()
	mux.DefaultMsgHandleFunc(defaultMsgHandler)
	mux.DefaultEventHandleFunc(defaultEventHandler)
	mux.MsgHandleFunc(request.MsgTypeText, textMsgHandler)
	mux.MsgHandleFunc(request.MsgTypeImage, imgMsgHandler)
	mux.EventHandleFunc(menu.EventTypeClick, menuClickEventHandler)

	msgHandler = mux
	msgServer = core.NewServer(config.WXOriID, config.WXAppID, config.WXToken, config.WXEncodeAESKey, msgHandler, nil)

	http.HandleFunc("/wx_callback", wxCallbackHandler)
	http.HandleFunc("/upload", upload)
	http.ListenAndServe(fmt.Sprintf(":%d", config.HTTPPort), nil)

}

func upload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	file, header, err := r.FormFile("pic")
	if err != nil {
		log.Error("%+v", err)
		return
	}
	defer file.Close()
	folder := "./data/pic/"
	tmp := "./data/tmp/"
	f, err := os.OpenFile(tmp+header.Filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Error("%+v", err)
		return
	}

	io.Copy(f, file)
	f.Close()
	path, _ := filepath.Abs(tmp + header.Filename)
	c := tesseract.ParseCard(path)
	if c != nil {
		os.Rename(tmp+header.Filename, folder+c.No+".jpg")
	}
	out, _ := json.Marshal(c)
	w.Write(out)
}
