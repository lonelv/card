package main

import (
	"fmt"
	"net/http"

	"github.com/skiplee85/card/conf"
	"github.com/skiplee85/card/dao"
	"github.com/skiplee85/card/route"
	"github.com/skiplee85/card/wx"
	"github.com/skiplee85/common/log"
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
	defaultHandler := route.GetRouteHandler(isDebug)
	http.ListenAndServe(fmt.Sprintf("%s:%d", conf.HTTP.Address, conf.HTTP.Port), defaultHandler)

}
