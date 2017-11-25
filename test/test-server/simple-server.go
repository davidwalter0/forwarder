package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/davidwalter0/go-cfg"
	"github.com/davidwalter0/go-tracer"
	"github.com/jehiah/go-strftime"
)

var err error
var trace = tracer.New()

var debugging = false

var done = make(chan bool)
var app App

// App application configuration struct
type App struct {
	Cert          string
	Key           string
	Host          string
	Port          string
	Path          string
	TraceDetailed bool
	TraceEnabled  bool
}

// Build version build string
var Build string

// Commit version commit string
var Commit string

// Now current formatted time
func Now() string {
	format := "%Y.%m.%d.%H.%M.%S.%z"
	now := time.Now()
	return strftime.Format(format, now)
}

var prefix = ""

func init() {
	if err = cfg.ProcessHoldFlags("APP", &app); err != nil {
		log.Fatalf("%v\n", err)
	}
	cfg.Freeze()
	array := strings.Split(os.Args[0], "/")
	me := array[len(array)-1]
	fmt.Println(me, "version built at:", Build, "commit:", Commit)
}

func run() {
	defer trace.Detailed(app.TraceDetailed).Enable(app.TraceEnabled).ScopedTrace()()
	validate()
	fmt.Println("PATH to serve  " + ":" + app.Path)
	fmt.Println("PORT on which  " + ":" + app.Port)
	fmt.Println("HOST interface " + ":" + app.Host)
	fmt.Printf("HTTPS/Listening on %s:%s and serving path: %s\n", app.Host, app.Port, app.Path)
	go vanillaServe(app.Host, app.Port)
	url := fmt.Sprintf("%s:%s", app.Host, app.Port)
	err := http.ListenAndServeTLS(url, app.Cert, app.Key, http.FileServer(http.Dir(app.Path)))
	if err != nil {
		log.Fatal(url, err)
	}
}

func vanillaServe(host, port string) {
	defer trace.Detailed(app.TraceDetailed).Enable(app.TraceEnabled).ScopedTrace()()
	p, _ := strconv.Atoi(port)
	p-- // http port = port - 1
	port = fmt.Sprintf("%d", p)
	fmt.Printf("HTTP /Listening on %s:%s and serving path: %s\n", app.Host, port, app.Path)
	listen := fmt.Sprintf("%s:%s", host, port)
	fmt.Println(http.ListenAndServe(listen, http.FileServer(http.Dir(app.Path))))
}

func validate() {
	if len(app.Path) == 0 || strings.Contains(app.Path, "..") {
		log.Fatal("APP_PATH not set or using .. to subvert app.Paths, using /target instead")
	} else {
		fmt.Println("APP_PATH=" + app.Path)
	}
	if len(app.Port) == 0 {
		log.Fatal("APP_PORT not set, using 8080")
	} else {
		fmt.Println("APP_PORT=" + app.Port)
	}
	if len(app.Host) == 0 {
		log.Fatal("APP_HOST not set, default bind all")
	} else {
		fmt.Println("APP_HOST=" + app.Host)
	}
}

func main() {
	defer trace.Detailed(app.TraceDetailed).Enable(app.TraceEnabled).ScopedTrace()()
	run()
}
