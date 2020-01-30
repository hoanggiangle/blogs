package util

import (
	"log"
	"reflect"
	"strings"
	"time"
)

type BuildInfo struct {
	Date     string
	Branch   string
	Revision string
}

func (b *BuildInfo) GetBranch() string {
	if b == nil {
		return ""
	}
	idx := strings.LastIndex(b.Branch, "/")
	return b.Branch[idx+1:]
}

func (b *BuildInfo) GetDate() time.Time {
	t, err := time.Parse(time.RFC3339, b.Date)
	if err != nil {
		log.Fatal("BuildInfo.Date:", err)
	}
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err == nil { // change timezone to HCM if possible
		t = t.In(loc)
	}
	return t
}

var buildInfo *BuildInfo

func SetBuildInfo(inf *BuildInfo) {
	if !reflect.DeepEqual(inf, &BuildInfo{}) {
		buildInfo = inf
	}
}

func GetBuildInfo() *BuildInfo {
	return buildInfo
}
