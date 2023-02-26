package app

import (
	"github.com/curltech/go-colla-biz/app/router"
	"github.com/curltech/go-colla-biz/app/websocket"
	base "github.com/curltech/go-colla-biz/base/controller"
	"github.com/curltech/go-colla-biz/controller"
	controller2 "github.com/curltech/go-colla-biz/rbac/controller"
	"github.com/curltech/go-colla-biz/view"
	"github.com/curltech/go-colla-core/config"
	"github.com/curltech/go-colla-core/logger"
	"github.com/kataras/iris/v12"
	irislogger "github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
)

var app *iris.Application

func setLog() {
	level, _ := config.GetString("log.level", "info")
	app.Logger().SetLevel(level)

	app.Use(irislogger.New())

	//f, _ := os.Create("iris.log")
	//app.Logger().SetOutput(f)
	//level := logger.Levels[logger.DebugLevel]
	//level.Name = "debug"         // default
	//level.Title = "[DBUG]"       // default
	//level.ColorCode = pio.Yellow // default
	//app.Logger().SetFormat("json", "    ")
	//app.Logger().SetLevelOutput("error", os.Stderr)
	//app.Logger().SetLevelFormat("debug", "json")
}

func newApp(name string) *iris.Application {
	app = iris.New().SetName(name)
	irisConfig := iris.YAML("conf/iris.yml")
	configurator := iris.WithConfiguration(irisConfig)
	app.Configure(configurator)
	// Optionally, add two built'n handlers
	// that can recover from any http-relative panics
	// and log the requests to the terminal.
	app.Use(recover.New())
	setLog()
	if config.AppParams.EnableSession {
		//会话控制器，每个请求都会经过会话控制器
		app.Use(base.SessionController)
	}
	if config.RbacParams.EnableCasbin {
		controller2.Set(app)
	}
	//app.Validator = validator.New()

	app.Favicon("./static/ico/favicon.ico")
	app.HandleDir("/js", iris.Dir("./static/js"))
	app.HandleDir("/assets", iris.Dir("./assets"))

	app.Use(iris.Compression)

	router.Set(app)
	view.Set(app)

	app.I18n.Load("./locales/*/*", "en-US", "el-GR", "zh-CN")
	app.I18n.SetDefault("zh-CN")

	return app
}

func Start() {
	if !config.AppParams.Enable {
		//启动空的主线程
		select {}
		return
	}
	start()
}

func start() {
	//主线程为iris应用
	appname := config.GetAppName()
	app := newApp(appname)
	config.AppParams.Name = appname
	controller.RegistTemplateParam("Title", config.AppParams.Name)
	websocket.Set(app)
	port := config.ServerParams.Port
	tlsmode := config.TlsParams.Mode
	if tlsmode != "none" {
		port = config.TlsParams.Port
	}

	logger.Sugar.Infof("successfully start iris app %v in port %v using %v tls mode,enjoy it!", appname, port, tlsmode)

	var irisAddr = ":" + port
	if config.ServerParams.Addr != "" {
		irisAddr = config.ServerParams.Addr + ":" + port
	}
	if tlsmode == "none" {
		app.Run(iris.Addr(irisAddr), iris.WithSocketSharding, iris.WithoutServerError(iris.ErrServerClosed))
	} else if tlsmode == "auto" {
		url := config.TlsParams.Url
		mail := config.TlsParams.Email
		app.Run(iris.AutoTLS(irisAddr, url, mail))
	} else if tlsmode == "cert" {
		cert := config.TlsParams.Cert
		key := config.TlsParams.Key
		app.Run(
			iris.TLS(irisAddr, cert, key),
		)
	}
}
