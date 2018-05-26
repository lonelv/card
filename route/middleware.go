package route

import (
	"fmt"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"

	"github.com/skiplee85/card/conf"
	"github.com/skiplee85/card/msg"
	"github.com/skiplee85/common/log"
)

func authMiddleware(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		log.Error("Authorization empty. %s %s", c.Request.Method, c.Request.URL)
		c.AbortWithStatusJSON(http.StatusUnauthorized, &msg.BaseResponse{Code: http.StatusUnauthorized})
		return
	}

	claims, code := parseToken(auth)
	if code == http.StatusOK {
		role := c.GetInt(keyRole)
		if role > claims.Role {
			c.AbortWithStatusJSON(http.StatusForbidden, &msg.BaseResponse{Code: http.StatusForbidden})
		} else {
			c.Set(keyUserClaims, claims)
		}
	} else {
		c.AbortWithStatusJSON(http.StatusOK, &msg.BaseResponse{Code: code})
	}
}

func parseToken(auth string) (*msg.UserClaims, int) {
	token, err := jwt.ParseWithClaims(auth, &msg.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method %v", token.Header["alg"])
		}
		return []byte(conf.Common.JWTSecret), nil
	})
	if err != nil {
		log.Error("Parse Authorization Fail. %s %s %+v", err)
		return nil, http.StatusUnauthorized
	}

	claims, ok := token.Claims.(*msg.UserClaims)
	if ok == false || token.Valid == false {
		return nil, http.StatusUnauthorized
	}
	return claims, http.StatusOK
}
