package wx

import (
	"fmt"
	"testing"

	"gopkg.in/chanxuehong/wechat.v2/mp/menu"
)

func TestDeleteMenu(t *testing.T) {
	appID := "wx98b04967828ab8b4"
	appSecret := "2d985ce80b2ff0412ff98ad27e5f8cbf"
	InitWXClient(appID, appSecret)

	fmt.Println(menu.Delete(wechatClient))
}
