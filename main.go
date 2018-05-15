package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-ini/ini"
	"github.com/skiplee85/card/dao"
	"github.com/skiplee85/card/log"
	"github.com/skiplee85/card/tesseract"
)

type Config struct {
	Level      string `ini:"LEVEL"`
	DBUser     string `ini:"DB_USER"`
	DBPassword string `ini:"DB_PASSWORD"`
	DBHost     string `ini:"DB_HOST"`
	DBPort     int    `ini:"DB_PORT"`
	DBDataBase string `ini:"DB_DATABASE"`
	HTTPPort   int    `ini:"HTTP_PORT"`
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

	http.HandleFunc("/upload", upload)
	http.ListenAndServe(":8080", nil)

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
