package appcmd

import (
	"github.com/urfave/cli"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	echohttp "github.com/labstack/echo/engine/standard"
	//echohttp "github.com/labstack/echo/engine/fasthttp"

	"strings"

	"config"
	log "logger"
	"fmt"
	"control"
	"models"
	"gopkg.in/macaron.v1"
	"github.com/go-macaron/binding"
	"net/http"
	"sync"
)

//TODO FIXME
var (
	_ = http.DefaultClient
)

var CmdWeb = cli.Command{
	Name:  "web",
	Usage: "Start Gogs web server",
	Description: `Gogs web server is the only thing you need to run,
				and it takes care of all the other things for you`,
	Action: runWeb,
	Flags: []cli.Flag{
		stringFlag("port", "12006", "Temporary port number to prevent conflict"),
		stringFlag("config", "config/app.ini", "Custom configuration file path"),
	},
}

func newEcho() *echo.Echo{
	e := echo.New()
	e.SetDebug(true)
	e.Use(middleware.Logger())
	//route
	group := e.Group("/account")
	//group.POST("/20006", control.Login)
	group.POST("/20006", control.Login)
	return e
}

var mutex sync.Mutex
var Count int

func AddCount(){
	mutex.Lock()
	Count ++
	fmt.Println("发起请求数：",Count)
	mutex.Unlock()
}
func newMacaron() *macaron.Macaron{
	m := macaron.New()
	//m.Use(macaron.Logger())
	m.Use(macaron.Renderer())
	m.Use(logger())


	//bindIgnErr := binding.BindIgnErr
	bind := binding.Bind
	//route
	m.Group("/account", func() {
		m.Combo("/20006").Get(bind(models.Accounts{}),control.LoginGet).Post(bind(models.Accounts{}),control.LoginGet)
	})
	return m
}

func logger() macaron.Handler {
	return func(ctx *macaron.Context) {
		AddCount()
		 //fmt.Println("接受：",ctx.Req.Header)
		 //fmt.Println("GET BODY：",ctx.Req.Request)
	}
}


func runWeb(ctx *cli.Context) error {




	// Flag for port number in case first time run conflict.
	if ctx.IsSet("port") {
		config.AppUrl = strings.Replace(config.AppUrl, config.HttpPort, ctx.String("port"), 1)
		config.HttpPort = ctx.String("port")
	}
	log.Info("Listen: %s", config.HttpPort)
	fmt.Println("正在监听端口:",config.HttpPort)
	//e := newEcho()
	//e.Run(echohttp.New(":"+config.HttpPort))
	_ = echohttp.NewRequest
	m := newMacaron()
	err := http.ListenAndServe(":"+config.HttpPort, m)
	if err != nil {
		log.Fatal(4, "Fail to start server: %v", err)
	}
	return nil
}