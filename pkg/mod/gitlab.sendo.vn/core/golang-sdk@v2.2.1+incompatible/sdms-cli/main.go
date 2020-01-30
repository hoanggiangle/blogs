// SDK Utils
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/template"
)

const PATHSEP string = string(os.PathSeparator)

const pkgName = "gitlab.sendo.vn/core/golang-sdk"
const pkgNativePath string = "gitlab.sendo.vn" + PATHSEP + "core" + PATHSEP + "golang-sdk"

var sNameRe = regexp.MustCompile(`^[[:alpha:]][[:alnum:]]*(-[[:alnum:]]+)*$`)

func init() {
	_, err := exec.LookPath("goimports")
	if err != nil {
		log.Fatal("goimports not found in PATH\n\n")
	}
	_, err = exec.LookPath("gofmt")
	if err != nil {
		log.Fatal("gofmt not found in PATH\n\n")
	}
}

func validateServiceName(sname string) bool {
	return sNameRe.MatchString(sname)
}

func writeTemplate(src, dst string, data interface{}) {
	if strings.Contains(src, ".DS_") {
		return
	}
	t := template.Must(template.ParseFiles(src))

	f, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	t.Execute(f, data)
}

func formatGo(file string) {
	cmd := exec.Command("goimports", "-w", file)
	err := cmd.Run()
	if err != nil {
		log.Fatalf("goimports %s: %s", file, err)
	}

	cmd = exec.Command("gofmt", "-w", file)
	err = cmd.Run()
	if err != nil {
		log.Fatalf("gofmt %s: %s", file, err)
	}
}

func findSdkPath() string {
	exe, _ := os.Executable()

	_, this_file, _, _ := runtime.Caller(0)

	paths := []string{
		path.Dir(path.Dir(this_file)),
		fmt.Sprintf("%s/../src/%s", filepath.Dir(exe), pkgName),
		filepath.Dir(exe) + "/..",
		"..",
	}
	for _, sdk := range paths {
		sdk, err := filepath.Abs(sdk)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := os.Stat(sdk); os.IsNotExist(err) {
			continue
		}
		if !strings.Contains(sdk, pkgNativePath) {
			continue
		}

		return sdk
	}

	log.Fatal("Can't find sdk path")
	return ""
}

func findTemplateDir(tmpl string) string {
	p, _ := filepath.Abs(fmt.Sprintf("%s/sdms-cli/%s-skeleton", findSdkPath(), tmpl))
	return p
}

func initTemplate(tmpl, sname, destpath string, force bool, resources map[string]bool) {
	destpath, _ = filepath.Abs(destpath)

	if !force {
		c := 0
		var checkEmpty = func(path string, info os.FileInfo, err error) error {
			c += 1
			if c > 1 {
				log.Fatal("Target is not empty")
			}
			return nil
		}
		filepath.Walk(destpath, checkEmpty)
	}

	tmplDir := findTemplateDir(tmpl)

	var createDir = func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if !info.IsDir() {
			return nil
		}

		target := destpath + strings.TrimPrefix(path, tmplDir)
		err = os.MkdirAll(target, 0755)
		if err != nil {
			log.Fatal(err)
		}

		return nil
	}
	filepath.Walk(tmplDir, createDir)

	sdkPath := findSdkPath()
	gopath := strings.Split(tmplDir, pkgNativePath)[0]
	gopath = strings.TrimRight(gopath, PATHSEP+"src"+PATHSEP)
	var data = struct {
		ServiceName string
		ImportPath  string
		GoPath      string
		SdkPath     string
		Resources   map[string]bool
	}{
		ServiceName: sname,
		GoPath:      gopath,
		SdkPath:     sdkPath,
		Resources:   resources,
	}

	{ // find import path
		f, err := os.OpenFile(destpath+"/test.go", os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			log.Fatal(err)
		}
		f.Write([]byte("package main"))
		f.Close()

		oldWd, _ := os.Getwd()
		os.Chdir(destpath)
		defer os.Chdir(oldWd)

		cmd := exec.Command("go", "list")
		out, err := cmd.Output()
		os.Remove(destpath + "/test.go")
		if err != nil {
			log.Fatal(err)
		}
		data.ImportPath = strings.TrimSpace(string(out))

		if strings.HasPrefix(data.ImportPath, "_") {
			log.Fatalf("Directory %s is not in GOPATH", destpath)
		}
	}

	var createFile = func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if info.IsDir() {
			return nil
		}

		target := destpath + strings.TrimPrefix(path, tmplDir)
		writeTemplate(path, target, data)

		if strings.HasSuffix(target, ".go") {
			formatGo(target)
		}

		return nil
	}
	filepath.Walk(tmplDir, createFile)
}

func mainUsage() {
	usage := fmt.Sprintf("Usage of %s:\n", os.Args[0])
	usage += "  init        Init an application template\n"
	usage += "  sdkpath     Show sendo-golang-sdk path\n"
	fmt.Fprint(os.Stderr, usage)
}

func main() {
	var tmpl string
	var force bool
	var sname string
	var path string
	var res string

	flag.StringVar(&tmpl, "t", "grpc", "Template to use (grpc, gin)")
	flag.BoolVar(&force, "f", false, "overwrite file if exists")
	flag.StringVar(&sname, "s", "", "(REQUIRED) Service name")
	flag.StringVar(&path, "p", "", "(REQUIRED) Destination path")
	flag.StringVar(&res, "r", "logfile,redis,mgo", "Resource to use: logfile, redis, mgo, amqp, consul")

	if len(os.Args) < 2 {
		mainUsage()
		os.Exit(2)
	}

	cmd := os.Args[1]
	switch cmd {
	case "init":
		flag.CommandLine.Parse(os.Args[2:])

		if sname == "" || path == "" {
			fmt.Fprintf(os.Stderr, "Usage of %s %s:\n", os.Args[0], cmd)
			flag.PrintDefaults()
			os.Exit(2)
		}

		if !validateServiceName(sname) {
			fmt.Println(`Service name must in format "text1[-text2[-text3]]"`)
			os.Exit(2)
		}

		resList := strings.Split(res, ",")
		resources := make(map[string]bool)
		for _, s := range resList {
			s = strings.TrimSpace(s)
			if s != "" {
				resources[s] = true
			}
		}

		initTemplate(tmpl, sname, path, force, resources)

	case "sdkpath":
		fmt.Println(findSdkPath())

	default:
		mainUsage()
		os.Exit(2)
	}
}
