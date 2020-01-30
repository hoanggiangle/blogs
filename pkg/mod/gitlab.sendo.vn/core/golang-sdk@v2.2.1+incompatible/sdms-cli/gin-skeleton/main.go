package main

import (
	"gitlab.sendo.vn/core/golang-sdk/util"

	"{{ .ImportPath }}/appsrc"
)

var (
	BUILD_DATE   string
	BUILD_BRANCH string
	BUILD_REV    string
)

func main() {
	util.SetBuildInfo(&util.BuildInfo{
		Date:     BUILD_DATE,
		Branch:   BUILD_BRANCH,
		Revision: BUILD_REV,
	})

	m := util.SimpleMain{}
	m.Add(&util.SimpleCommand{
		Name: "run",
		Desc: "Run " + appsrc.SERVICE_NAME + " service",
		Func: func(args []string) { appsrc.CreateApp(args, "").Run() },
	})
	m.Add(&util.SimpleCommand{
		Name: "outenv",
		Desc: "output all environment variables",
		Func: func(args []string) { appsrc.CreateApp(args, "").OutputEnv() },
	})

	m.Execute()
}
