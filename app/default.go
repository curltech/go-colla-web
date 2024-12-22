package app

import (
	"github.com/curltech/go-colla-core/config"
	"github.com/curltech/go-colla-core/logger"
	"github.com/curltech/go-colla-web/app/router"
	"github.com/curltech/go-colla-web/app/websocket"
	base "github.com/curltech/go-colla-web/base/controller"
	"github.com/curltech/go-colla-web/controller"
	controller2 "github.com/curltech/go-colla-web/rbac/controller"
	"github.com/kataras/iris/v12"
	irislogger "github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
)

var app *iris.Application

/*
*
设置视图
*/
func Set(app *iris.Application) {
	if config.AppParams.Template == "html" {
		tmpl := iris.HTML("./view", ".html")
		tmpl.Reload(true)
		app.RegisterView(tmpl)
	} else if config.AppParams.Template == "pug" {
		tmpl := iris.Pug("./view", ".pug")
		tmpl.Reload(true)
		app.RegisterView(tmpl)
	}
}

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
	app.HandleDir("/assets", iris.Dir("./static/assets"))
	app.HandleDir("/ico", iris.Dir("./static/ico"))
	app.HandleDir("/icon", iris.Dir("./static/icon"))
	app.HandleDir("/icons", iris.Dir("./static/icons"))
	app.HandleDir("/css", iris.Dir("./static/css"))
	app.HandleDir("/image", iris.Dir("./static/image"))
	app.HandleDir("/images", iris.Dir("./static/images"))
	app.HandleDir("/img", iris.Dir("./static/img"))
	app.HandleDir("/font", iris.Dir("./static/font"))
	app.HandleDir("/fonts", iris.Dir("./static/fonts"))

	app.Use(iris.Compression)

	router.Set(app)
	Set(app)

	err := app.I18n.Load("./locales/*/*", "en-US", "el-GR", "zh-CN")
	if err != nil {
		return nil
	}
	app.I18n.SetDefault("zh-CN")

	return app
}

func Start() {
	if !config.AppParams.Enable {
		//启动空的主线程
		select {}
	}
	start()
}

func start() {
	//主线程为iris应用
	appname := config.GetAppName()
	app := newApp(appname)
	config.AppParams.Name = appname
	err := controller.RegistTemplateParam("Title", config.AppParams.Name)
	if err != nil {
		return
	}
	websocket.Set(app)
	port := config.ServerParams.Port
	tlsmode := config.TlsParams.Mode
	if tlsmode != "none" {
		port = config.TlsParams.Port
	}

	var irisAddr = ":" + port
	if config.ServerParams.Addr != "" {
		irisAddr = config.ServerParams.Addr + ":" + port
	}
	logger.Sugar.Infof("Successfully start iris app %v in port %v using %v tls mode,enjoy it!", appname, port, tlsmode)
	if tlsmode == "none" {
		err := app.Run(iris.Addr(irisAddr), iris.WithSocketSharding, iris.WithoutServerError(iris.ErrServerClosed))
		if err != nil {
			logger.Sugar.Errorf("Start iris app failure %v", err)
			return
		}
	} else if tlsmode == "auto" {
		url := config.TlsParams.Url
		mail := config.TlsParams.Email
		err := app.Run(iris.AutoTLS(irisAddr, url, mail))
		if err != nil {
			logger.Sugar.Errorf("Start iris app failure %v", err)
			return
		}
	} else if tlsmode == "cert" {
		cert := config.TlsParams.Cert
		key := config.TlsParams.Key
		err := app.Run(
			iris.TLS(irisAddr, cert, key),
		)
		if err != nil {
			logger.Sugar.Errorf("Start iris app failure %v", err)
			return
		}
	}
}
