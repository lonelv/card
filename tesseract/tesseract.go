package tesseract

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/skiplee85/card/dao"
	"github.com/skiplee85/card/log"
)

func execShell(s string) []string {
	cmd := exec.Command("/bin/bash", "-c", s)
	var out bytes.Buffer
	var errMsg bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &errMsg
	err := cmd.Run()
	if err != nil {
		log.Error("Shell exec error. %+v", err)
	}
	ret := strings.Split(out.String(), "\n")
	if len(ret) > 0 {
		ret = ret[:len(ret)-1]
	}
	return ret
}

// ParseCard 获取卡信息
func ParseCard(path string) *dao.Card {
	cmd := fmt.Sprintf("tesseract %s stdout -l eng+chi_sim | sed -e 's/[^0-9]//g' -e '/^[[:space:]]*$/d' -e '/^.\\{1,16\\}$/d'", path)
	ret := execShell(cmd)
	if len(ret) == 0 || len(ret)%2 != 0 {
		log.Error("Error parse. %+v\n", ret)
		return nil
	}

	if len(ret[0]) != 18 {
		log.Error("Error parse. Secret length must 18, got %d. %+v\n", len(ret[0]), ret)
		return nil
	}
	if len(ret[1]) != 17 {
		log.Error("Error parse. No length must 17, got %d. %+v\n", len(ret[1]), ret)
		return nil
	}
	c := &dao.Card{
		No:     ret[1],
		Secret: ret[0],
		Create: time.Now(),
	}
	dao.MgoExecCard(func(sc *mgo.Collection) {
		f := &dao.Card{}
		err := sc.Find(bson.M{"no": c.No}).One(f)
		if err == nil {
			sc.Update(bson.M{"no": c.No}, bson.M{"$set": bson.M{"secret": c.Secret}})
		} else {
			if err == mgo.ErrNotFound {
				sc.Insert(c)
			} else {
				log.Error("Mongo Error", err)
			}
		}
	})
	log.Release("Parse Succ! No:%s\nSecret:%s\n", c.No, c.Secret)

	return c
}
