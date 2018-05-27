package route

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/cors"

	"github.com/skiplee85/card/msg"
	"github.com/skiplee85/common/log"
)

const (
	keyUserClaims = "keyUserClaims"
	keyRole       = "keyRole"
)

var errMsg = map[int]string{}

// Context gin.Context 二次封装
type Context struct {
	*gin.Context
}

type baseRoute struct {
	Method  string
	Path    string
	Handler func(*Context)
	Role    int // 访问所需的最小角色，高级角色拥有低级角色的权限。0 表示没要求，未授权的用户也可以访问。
	Child   []*baseRoute
}

// GetRouteHandler 获取路由
func GetRouteHandler(isDebug bool) http.Handler {
	if isDebug {
		gin.SetMode(gin.ReleaseMode)
	}
	defaultRoute := gin.Default()
	for _, r := range routeConf {
		createRouteHandler(r, &defaultRoute.RouterGroup)
	}

	// 跨域请求
	c := cors.New(cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "*"},
		AllowedHeaders:   []string{"Authorization", "*"},
		AllowCredentials: true,
		Debug:            false,
	})

	return c.Handler(defaultRoute)
}

func createRouteHandler(rConf *baseRoute, g *gin.RouterGroup) {
	r := *rConf
	hs := []gin.HandlerFunc{}
	if r.Role > 0 {
		hs = append(hs, getRoleMiddleware(r.Role), authMiddleware)
	}
	// group
	if len(r.Child) > 0 {
		gg := g.Group(r.Path)
		for _, rr := range r.Child {
			createRouteHandler(rr, gg)
		}
	} else {
		h := func(c *gin.Context) {
			r.Handler(&Context{Context: c})
		}
		hs = append(hs, h)
		g.Handle(r.Method, r.Path, hs...)
	}

}

// ValidaArgs 检查参数
func (c *Context) ValidaArgs(args interface{}) error {
	if err := c.ShouldBind(args); err != nil {
		log.Error("invalid args. %v. url=%s", err.Error(), c.Request.URL.String())
		c.AbortWithStatusJSON(http.StatusOK, &msg.BaseResponse{Code: msg.ERROR_INVALID_ARGUMENTS, Msg: err.Error()})
		return err
	}
	return nil
}

// SendError 下发错误
func (c *Context) SendError(code int) {
	httpStatus := http.StatusInternalServerError
	if code >= 1000 {
		httpStatus = http.StatusBadRequest
	}
	c.AbortWithStatusJSON(httpStatus, &msg.BaseResponse{
		Code: code,
		Msg:  errMsg[code],
	})
}

// Send 下发消息
func (c *Context) Send(data interface{}) {
	c.JSON(http.StatusOK, &msg.BaseResponse{
		Data: data,
	})
}

// SendWithPagination 带翻页信息
func (c *Context) SendWithPagination(data interface{}, p *msg.Pagination) {
	c.JSON(http.StatusOK, &msg.BaseResponse{
		Data:       data,
		Pagination: p,
	})
}

// GetClaims 获取JWT信息包
func (c *Context) GetClaims() *msg.UserClaims {
	v, ok := c.Get(keyUserClaims)
	if !ok {
		return nil
	}
	return v.(*msg.UserClaims)
}

// GetIP 获取IP
func (c *Context) GetIP() string {
	ip := ""
	findHeader := []string{
		"X-Forwarded-For",
		"Proxy-Client-IP",
		"WL-Proxy-Client-IP",
		"HTTP_CLIENT_IP",
		"HTTP_X_FORWARDED_FOR",
	}
	for _, h := range findHeader {
		ip = c.GetHeader(h)
		if ip != "" {
			break
		}
	}
	if ip == "" {
		ips := strings.Split(c.Request.RemoteAddr, ":")
		ip = ips[0]
	}
	return ip
}

// Finish 完成请求
func (c *Context) Finish(data interface{}, eno int) {
	if eno != msg.RET_OK {
		c.SendError(eno)
		return
	}
	c.Send(data)
}
