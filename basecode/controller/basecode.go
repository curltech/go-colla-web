package controller

import (
	"github.com/curltech/go-colla-biz/basecode/entity"
	"github.com/curltech/go-colla-biz/basecode/service"
	"github.com/curltech/go-colla-biz/controller"
	"github.com/curltech/go-colla-core/container"
	"github.com/curltech/go-colla-core/util/message"
	"github.com/kataras/iris/v12"
)

/**
控制层代码需要做数据转换，调用服务层的代码，由于数据转换的结构不一致，因此每个实体（外部rest方式访问）的控制层都需要写一遍
*/
type BaseCodeController struct {
	controller.BaseController
}

var baseCodeController *BaseCodeController

func (this *BaseCodeController) ParseJSON(json []byte) (interface{}, error) {
	var entities = make([]*entity.BaseCode, 0)
	err := message.Unmarshal(json, &entities)

	return &entities, err
}

func (this *BaseCodeController) GetBaseCode(ctx iris.Context) {
	baseCode := &entity.BaseCode{}
	err := ctx.ReadJSON(baseCode)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())

		return
	}
	if baseCode.BaseCodeId != "" {
		baseCode, err := service.GetBaseCodeService().GetBaseCode(baseCode.BaseCodeId)
		if err != nil {
			ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())

			return
		}
		ctx.JSON(baseCode)
	} else {
		ctx.StopWithJSON(iris.StatusInternalServerError, "BlankBaseCodeId")
	}
}

/**
注册bean管理器，注册序列
*/
func init() {
	baseCodeController = &BaseCodeController{
		BaseController: controller.BaseController{
			BaseService: service.GetBaseCodeService(),
		},
	}
	baseCodeController.BaseController.ParseJSON = baseCodeController.ParseJSON
	container.RegistController("baseCode", baseCodeController)
}
