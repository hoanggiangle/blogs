package tests

import (
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	sdms "gitlab.sendo.vn/core/golang-sdk"
	"gitlab.sendo.vn/core/golang-sdk/sgin"
	"gitlab.sendo.vn/core/golang-sdk/ssd"
)

func registerHandlers(r *gin.Engine) {
	r.GET("/ping", pingHandler)
}

func pingHandler(c *gin.Context) {
	c.String(200, "pong")
}

func gin_request(port int) {
	url := fmt.Sprintf("http://127.0.0.1:%d/ping", port)

	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(resp)
}

func TestGin(t *testing.T) {
	app := sdms.NewApp(&sdms.AppConfig{
		Args: []string{
			"-port", "0",
			"-addr", "127.0.0.1",
			"-gin-mode", "release",
		},
		UseNewFlagSet: true,
	})

	cnf := sgin.Config{
		App:         app,
		SD:          ssd.NewNullConsulSD(),
		ServiceName: "test-gin",
		RegFunc:     registerHandlers,
	}
	gSvc, _ := sgin.New(&cnf)
	app.RegMainService(gSvc)

	defer executeApp(app)()

	gin_request(gSvc.Port())
}
