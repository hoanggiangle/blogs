package appsrc

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/gin-gonic/gin"

	"{{ .ImportPath }}/appsrc/models"
	"{{ .ImportPath }}/appsrc/resprovider"
)

func registerHandlers(r *gin.Engine) {
	r.GET("/ping", pingHandler)
	r.GET("/ip", clientIpHandler)
	{{- if .Resources.redis }}
	r.GET("/redis", redisHandler)
	{{- end }}
	{{- if .Resources.mgo }}
	r.GET("/mongo", mongoHandler)
	{{- end }}
}

func pingHandler(c *gin.Context) {
	c.String(200, "pong\n")
}

func clientIpHandler(c *gin.Context) {
	c.String(200, c.ClientIP()+"\n")
}

{{- if .Resources.redis }}

func countReq() (int, error) {
	r := resprovider.GetInstance().Redis()
	defer r.Close()

	// sample use redis
	return redis.Int(r.Do("INCR", "visit-count"))
}

func redisHandler(c *gin.Context) {
	rp := resprovider.GetInstance()
	log := rp.Logger("countHandler")

	count, err := countReq()
	if err != nil {
		log.Error(err)
		c.String(500, "DB error")
		return
	}

	data := fmt.Sprintf(`Visit count: %d\n`, count)
	c.String(200, data+"\n")
}
{{- end }}

{{- if .Resources.mgo }}

func mongoHandler(c *gin.Context) {
	rp := resprovider.GetInstance()
	log := rp.Logger("countHandler")

	// sample use mongo
	vm := &models.VisitorModel{}
	err := vm.Insert(&models.Visitor{c.ClientIP(), c.Request.UserAgent(), time.Now()})
	if err != nil {
		log.Error(err)
		c.String(500, err.Error())
		return
	}

	var vList []models.Visitor
	if err := vm.Last10(&vList); err != nil {
		log.Error(err)
		c.String(500, err.Error())
		return
	}
	data := "Last 10 visitors:\n"
	for _, v := range vList {
		data += fmt.Sprintf(" - %v\n", v)
	}
	c.String(200, data+"\n")
}
{{- end }}
