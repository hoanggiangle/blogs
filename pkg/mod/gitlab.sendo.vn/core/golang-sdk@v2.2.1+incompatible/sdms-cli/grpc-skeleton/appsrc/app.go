package appsrc

import (
	"os"
	"strings"

	sdms "gitlab.sendo.vn/core/golang-sdk"
	"gitlab.sendo.vn/core/golang-sdk/sgrpc"
	"gitlab.sendo.vn/core/golang-sdk/ssd"

	"{{ .ImportPath }}/appsrc/resprovider"
)

const SERVICE_NAME = "{{ .ServiceName }}"

func CreateApp(args []string, job string) sdms.Application {
	name := ""
	if len(os.Args) >= 2 {
		name = strings.Join(os.Args[:2], " ")
	}
	app := sdms.NewApp(&sdms.AppConfig{
		Name: name,
		Args: args,
	})

	sd := ssd.NewConsul(app)
	app.RegService(sd)

	resprovider.InitResourceProvider(app)

	var main sdms.RunnableService

	switch job {
	default:
		cnf := sgrpc.GrpcConfig{
			App:          app,
			SD:           sd,
			ServiceName:  SERVICE_NAME,
			RegisterFunc: registerServices,
		}
		main = sgrpc.New(&cnf)
	}

	app.RegMainService(main)

	return app
}
