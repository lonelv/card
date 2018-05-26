package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"gopkg.in/mgo.v2"

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
	http.HandleFunc("/save-card", saveCard)
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
	var ss = map[string]string{}
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &ss)
	if err != nil {
		log.Error("%s\nParse Error:%+v", body, err)
		return
	}
	openid := ss["openid"]
	data := ss["data"]
	wx.SendNoticeImgBase64(openid, data)
}

func sendImgURL(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	openid := query["openid"][0]
	url := query["url"][0]
	wx.SendNoticeImgURL(openid, url)
}

func saveCard(w http.ResponseWriter, r *http.Request) {
	var err error
	r.ParseForm()
	no := r.FormValue("no")
	s := r.FormValue("secret")
	data := r.FormValue("data")

	bs, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		errMsg := fmt.Sprintf("Not base64 img. Error: %v", err)
		log.Error(errMsg)
		w.Write([]byte(errMsg))
		return
	}
	f := fmt.Sprintf("./data/pic/%s.jpg", no)

	file, err := os.OpenFile(f, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Error("%+v", err)
		return
	}

	file.Write(bs)
	file.Close()

	c := &dao.Card{
		No:     no,
		Secret: s,
		Create: time.Now(),
	}
	dao.MgoExecCard(func(sc *mgo.Collection) {
		_, err = sc.Upsert(bson.M{"no": c.No}, c)
	})
	if err != nil {
		errMsg := fmt.Sprintf("Mongo Error.%v", err)
		log.Error(errMsg)
		w.Write([]byte(errMsg))
	} else {
		w.Write([]byte("Success!"))
	}
}
