package main

import (
	"github.com/curltech/go-colla-biz/app"
	_ "github.com/curltech/go-colla-core/cache"
	/**
	  引入包定义，执行对应包的init函数，从而引入某功能，在init函数根据初始化参数配置觉得是否启动该功能
	*/
	_ "github.com/curltech/go-colla-core/content"
	_ "github.com/curltech/go-colla-core/repository/search"
)

func main() {
	app.Start()
}
