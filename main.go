package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-ini/ini"
	"github.com/skiplee85/card/dao"
	"github.com/skiplee85/card/log"
	"github.com/skiplee85/card/tesseract"
	"github.com/skiplee85/card/wx"
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
	wx.InitWXServer(config.WXAppID, config.WXToken, config.WXOriID, config.WXEncodeAESKey)
	wx.InitWXClient(config.WXAppID, config.WXAppSecret)

	http.HandleFunc("/upload", upload)
	http.HandleFunc("/send-imgbase64", sendImgBase64)
	http.HandleFunc("/send-imgurl", sendImgURL)
	http.ListenAndServe(fmt.Sprintf(":%d", config.HTTPPort), nil)

}

func upload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	file, _, err := r.FormFile("pic")
	if err != nil {
		log.Error("%+v", err)
		return
	}
	defer file.Close()
	imgByte, _ := ioutil.ReadAll(file)
	c := tesseract.ParseCardByBytes(imgByte)
	if c != nil {
		out, _ := json.Marshal(c)
		w.Write(out)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("解析失败，请重拍~"))
	}

}

func sendImgBase64(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	openid := query["openid"][0]
	data := query["data"][0]
	wx.SendNoticeImgBase64(openid, data)
}

func sendImgURL(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	openid := query["openid"][0]
	url := query["url"][0]
	wx.SendNoticeImgURL(openid, url)
}
