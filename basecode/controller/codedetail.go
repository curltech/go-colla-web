package controller

import (
	"github.com/curltech/go-colla-biz/basecode/entity"
	"github.com/curltech/go-colla-biz/basecode/service"
	"github.com/curltech/go-colla-biz/controller"
	"github.com/curltech/go-colla-core/container"
	"github.com/curltech/go-colla-core/util/message"
)

/**
控制层代码需要做数据转换，调用服务层的代码，由于数据转换的结构不一致，因此每个实体（外部rest方式访问）的控制层都需要写一遍
*/
type CodeDetailController struct {
	controller.BaseController
}

var codeDetailController *CodeDetailController

func (this *CodeDetailController) ParseJSON(json []byte) (interface{}, error) {
	var entities = make([]*entity.CodeDetail, 0)
	err := message.Unmarshal(json, &entities)

	return &entities, err
}

/**
注册bean管理器，注册序列
*/
func init() {
	codeDetailController = &CodeDetailController{
		BaseController: controller.BaseController{
			BaseService: service.GetCodeDetailService(),
		},
	}
	codeDetailController.BaseController.ParseJSON = codeDetailController.ParseJSON
	container.RegistController("codeDetail", codeDetailController)
}
