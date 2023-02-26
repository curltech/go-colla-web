package controller

import (
	"github.com/curltech/go-colla-biz/controller"
	"github.com/curltech/go-colla-biz/ruleengine/entity"
	service2 "github.com/curltech/go-colla-biz/ruleengine/service"
	"github.com/curltech/go-colla-core/container"
	"github.com/curltech/go-colla-core/util/message"
)

/**
控制层代码需要做数据转换，调用服务层的代码，由于数据转换的结构不一致，因此每个实体（外部rest方式访问）的控制层都需要写一遍
*/
type RuleDefinitionController struct {
	controller.BaseController
}

var ruleDefinitionController *RuleDefinitionController

func (this *RuleDefinitionController) ParseJSON(json []byte) (interface{}, error) {
	var entities = make([]*entity.RuleDefinition, 0)
	err := message.Unmarshal(json, &entities)

	return &entities, err
}

/**
注册bean管理器，注册序列
*/
func init() {
	ruleDefinitionController = &RuleDefinitionController{
		BaseController: controller.BaseController{
			BaseService: service2.GetRuleDefinitionService(),
		},
	}
	ruleDefinitionController.BaseController.ParseJSON = ruleDefinitionController.ParseJSON
	container.RegistController("ruleDefinition", ruleDefinitionController)
}
