package router

import (
	"github.com/curltech/go-colla-biz/controller"
	controller2 "github.com/curltech/go-colla-biz/rbac/controller"
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
)

/**
设置路由表
1.根路径对应/login.html，由LoginController处理
2.自动路由，根据方法名分派
3.html后缀的页面，采用pug模版处理
*/
func Set(app *iris.Application) {
	//crs := cors.New(cors.Options{
	//	AllowedOrigins:   []string{"*"},   //允许通过的主机名称
	//	AllowCredentials: true,
	//})
	crs := cors.AllowAll()

	//handler, err := newrelic.New(newrelic.ConfigAppName("APP_SERVER_NAME"), newrelic.ConfigLicense("NEWRELIC_LICENSE_KEY"))
	//if err != nil {
	//	logger.Sugar.Errorf("%v", err)
	//}
	//app.Use(handler)
	//
	//p := prometheusMiddleware.New("serviceName", 300, 1200, 5000)
	//app.Use(p.ServeHTTP)
	//app.Get("/metrics", iris.FromStd(p.ServeHTTP))

	//注册页面处理控制器

	app.Any("/{name:string suffix(.html)}", crs, controller2.Protected, controller.HTMLController)

	//Method:   POST
	//Resource: http://localhost:8080/receive
	app.Options("/receive", crs, controller2.Protected)
	//app.Any("/receive", crs, controller2.Protected, controller.ReceiveController) // ReceiveController or ReceivePCController

	// Method:   Post
	// Resource: http://localhost:8080/user/add，调用UserController.Add
	app.Options("/{serviceName:string}/{methodName:string}", crs, controller2.Protected, controller.MainController)
	app.Any("/{serviceName:string}/{methodName:string}", crs, controller2.Protected, controller.MainController)

	app.Options("/upload", crs, controller2.Protected, controller.UploadController)
	app.Any("/upload", crs, controller2.Protected, controller.UploadController)

	app.Options("/download", crs, controller2.Protected, controller.DownloadController)
	app.Any("/download", crs, controller2.Protected, controller.DownloadController)

	app.Get("/", func(ctx iris.Context) {
		ctx.View("index", controller.TemplateParams)
	})
}
