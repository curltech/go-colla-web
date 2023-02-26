package controller

import (
	"github.com/curltech/go-colla-biz/actual"
	"github.com/curltech/go-colla-biz/actual/entity"
	"github.com/curltech/go-colla-biz/actual/service"
	"github.com/curltech/go-colla-biz/controller"
	"github.com/curltech/go-colla-biz/spec"
	"github.com/curltech/go-colla-core/container"
	baseerror "github.com/curltech/go-colla-core/error"
	"github.com/curltech/go-colla-core/util/message"
	"github.com/kataras/iris/v12"
)

/**
控制层代码需要做数据转换，调用服务层的代码，由于数据转换的结构不一致，因此每个实体（外部rest方式访问）的控制层都需要写一遍
*/
type RoleController struct {
	controller.BaseController
}

var roleController *RoleController

func (this *RoleController) ParseJSON(json []byte) (interface{}, error) {
	var entities = make([]*entity.Role, 0)
	err := message.Unmarshal(json, &entities)

	return &entities, err
}

func (this *RoleController) Create(ctx iris.Context) {
	condiBean := entity.Role{}
	err := ctx.ReadJSON(&condiBean)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
	roleSpec, _ := spec.GetMetaDefinition().GetRoleSpec(condiBean.SpecId, condiBean.EffectiveDate)
	role := actual.Create(condiBean.SchemaName, roleSpec)
	ctx.JSON(role)
}

func (this *RoleController) Save(ctx iris.Context) {
	condiBean := entity.Role{}
	err := ctx.ReadJSON(&condiBean)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
	num, err := actual.Save(condiBean.SchemaName, condiBean.Id)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	} else {
		ctx.JSON(num)
	}
}

func (this *RoleController) Load(ctx iris.Context) {
	condiBean := entity.Role{}
	err := ctx.ReadJSON(&condiBean)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
	role := actual.Load(condiBean.SchemaName, condiBean.Id)
	ctx.JSON(role)
}

func (this *RoleController) Version(ctx iris.Context) {
	condiBean := entity.Role{}
	err := ctx.ReadJSON(&condiBean)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
	role := actual.Load(condiBean.SchemaName, condiBean.Id)
	if role != nil {
		role = role.Version()
		ctx.JSON(role)
	}
}

func (this *RoleController) Find(ctx iris.Context) {
	condiBean := entity.Role{}
	err := ctx.ReadJSON(&condiBean)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
	role := actual.Get(condiBean.SchemaName, condiBean.Id)
	if role != nil {
		value, _, ok := role.Find(condiBean.Path)
		if ok {
			ctx.JSON(value)
		} else {
			ctx.StopWithJSON(iris.StatusInternalServerError, baseerror.Error_NoValue)
		}
	} else {
		ctx.StopWithJSON(iris.StatusInternalServerError, baseerror.Error_NotFound)
	}
}

func (this *RoleController) GetActual(ctx iris.Context) {
	condiBean := entity.Role{}
	err := ctx.ReadJSON(&condiBean)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
	role := actual.Load(condiBean.SchemaName, condiBean.Id)
	if role != nil {
		actual := role.GetActual(false)
		ctx.JSON(actual)
	}
}

/**
注册bean管理器，注册序列
*/
func init() {
	roleController = &RoleController{
		BaseController: controller.BaseController{
			BaseService: service.GetRoleService(),
		},
	}
	roleController.BaseController.ParseJSON = roleController.ParseJSON
	container.RegistController("atlRole", roleController)
}
