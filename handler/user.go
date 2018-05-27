package handler

import (
	"github.com/skiplee85/card/msg"
	"github.com/skiplee85/card/service/user"
	"github.com/skiplee85/common/log"
	"github.com/skiplee85/common/route"
)

func login(c *route.Context) {
	var req msg.LoginReq
	if err := c.ValidaArgs(&req); err != nil {
		return
	}

	user, ret := user.Login(req.Username, req.Password, c.GetIP())
	if ret != msg.RET_OK {
		c.SendError(ret)
		return
	}

	log.Debug("User [%d %s] login", user.UserID, user.Username)
	c.Send(user)
}
