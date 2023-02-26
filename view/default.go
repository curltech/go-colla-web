package view

import (
	"github.com/curltech/go-colla-core/config"
	"github.com/kataras/iris/v12"
)

/**
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
