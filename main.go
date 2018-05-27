package main

import (
	"fmt"
	"net/http"

	"github.com/skiplee85/card/conf"
	"github.com/skiplee85/card/dao"
	"github.com/skiplee85/card/handler"
	"github.com/skiplee85/card/wx"
	"github.com/skiplee85/common/log"
	"github.com/skiplee85/common/route"
)

func main() {

	log.InitLog(conf.Common.Level)
	dao.InitMongo()
	wx.InitWXServer()
	wx.InitWXClient()

	isDebug := false
	if conf.Common.Level == "Debug" {
		isDebug = true
	}
	defaultHandler := route.GetRouteHandler(handler.RouteConf, conf.Common.JWTSecret, isDebug)
	http.Handle("/", defaultHandler)
	http.ListenAndServe(fmt.Sprintf("%s:%d", conf.HTTP.Address, conf.HTTP.Port), nil)

}
