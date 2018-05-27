package handler

import (
	"github.com/skiplee85/card/msg"
	"github.com/skiplee85/card/service/card"
	"github.com/skiplee85/common/route"
)

func listCard(c *route.Context) {
	var req msg.ListCardReq
	if err := c.ValidaArgs(&req); err != nil {
		return
	}

	c.SendWithPagination(card.List(req))
}

func modifyCard(c *route.Context) {
	var req msg.ModifyCardReq
	if err := c.ValidaArgs(&req); err != nil {
		return
	}

	c.Finish(card.Modify(req))

}

func saveCard(c *route.Context) {
	var req msg.SaveCardReq
	if err := c.ValidaArgs(&req); err != nil {
		return
	}

	c.Finish(card.Save(req))
}

func getCardData(c *route.Context) {
	no, isExist := c.GetQuery("no")
	if !isExist {
		c.SendError(msg.ERROR_REQUEST)
	}
	c.Finish(card.GetData(no))
}
