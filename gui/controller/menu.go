package controller

import (
	"github.com/curltech/go-colla-biz/controller"
	"github.com/curltech/go-colla-biz/gui/entity"
	"github.com/curltech/go-colla-biz/gui/service"
	"github.com/curltech/go-colla-core/container"
	"github.com/curltech/go-colla-core/util/message"
	"github.com/kataras/iris/v12"
)

/**
控制层代码需要做数据转换，调用服务层的代码，由于数据转换的结构不一致，因此每个实体（外部rest方式访问）的控制层都需要写一遍
*/
type GuiMenuController struct {
	controller.BaseController
}

var guiMenuController *GuiMenuController

func (this *GuiMenuController) ParseJSON(json []byte) (interface{}, error) {
	var entities = make([]*entity.GuiMenu, 0)
	err := message.Unmarshal(json, &entities)

	return &entities, err
}

func (this *GuiMenuController) GetGuiMenu(ctx iris.Context) {
	guiMenu := &entity.GuiMenu{}
	err := ctx.ReadJSON(guiMenu)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())

		return
	}
	menu, err := service.GetGuiMenuService().GetGuiMenu(guiMenu.MenuId)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())

		return
	}
	ctx.JSON(menu)
}

/**
注册bean管理器，注册序列
*/
func init() {
	guiMenuController = &GuiMenuController{
		BaseController: controller.BaseController{
			BaseService: service.GetGuiMenuService(),
		},
	}
	guiMenuController.BaseController.ParseJSON = guiMenuController.ParseJSON
	container.RegistController("guiMenu", guiMenuController)
}
