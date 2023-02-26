package controller

import (
	"fmt"
	"github.com/curltech/go-colla-biz/controller"
	"github.com/curltech/go-colla-biz/spec"
	"github.com/curltech/go-colla-biz/spec/entity"
	service2 "github.com/curltech/go-colla-biz/spec/service"
	"github.com/curltech/go-colla-core/container"
	"github.com/curltech/go-colla-core/util/convert"
	"github.com/curltech/go-colla-core/util/message"
	"github.com/kataras/iris/v12"
	"os"
	"time"
)

/**
控制层代码需要做数据转换，调用服务层的代码，由于数据转换的结构不一致，因此每个实体（外部rest方式访问）的控制层都需要写一遍
*/
type RoleSpecController struct {
	controller.BaseController
}

var roleSpecController *RoleSpecController

func (this *RoleSpecController) ParseJSON(json []byte) (interface{}, error) {
	var entities = make([]*entity.RoleSpec, 0)
	err := message.Unmarshal(json, &entities)

	return &entities, err
}

func (this *RoleSpecController) Upload(ctx iris.Context) {
	filename := "/Users/hujingsong/Downloads/数据模型梳理V6.8.7.xlsx"
	_, err := os.Stat(filename)
	if err == nil || os.IsExist(err) {
		spec.UploadExcel(filename)
	}
}

func (this *RoleSpecController) GetMetaDefinition(ctx iris.Context) {
	md := spec.GetMetaDefinition()
	ctx.JSON(md)
}

func (this *RoleSpecController) GetRoleSpec(ctx iris.Context) {
	params := make(map[string]interface{})
	err := ctx.ReadJSON(&params)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())

		return
	}
	var specId uint64
	v, ok := params["specId"]
	if ok {
		v, err := convert.ToObject(fmt.Sprintf("%v", v), "uint64")
		if err != nil {
			ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())

			return
		}
		specId = v.(uint64)
	} else {
		ctx.StopWithJSON(iris.StatusInternalServerError, "NoSpecId")

		return
	}
	var effectiveDate *time.Time = nil
	v, ok = params["effectiveDate"]
	if ok {
		v, err := convert.ToObject(fmt.Sprintf("%v", v), "time.Time")
		if err != nil {
			ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())

			return
		}
		t := v.(time.Time)
		effectiveDate = &t
	}

	roleSpec, modelSpec := spec.GetMetaDefinition().GetRoleSpec(specId, effectiveDate)
	if roleSpec == nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, "NotFound")

		return
	}

	var mode string
	v, ok = params["mode"]
	if ok {
		mode, _ = v.(string)
	}
	if mode == "roleSpec" {
		ctx.JSON(roleSpec)
	} else {
		ctx.JSON(modelSpec)
	}
}

/**
注册bean管理器，注册序列
*/
func init() {
	roleSpecController = &RoleSpecController{
		BaseController: controller.BaseController{
			BaseService: service2.GetRoleSpecService(),
		},
	}
	roleSpecController.BaseController.ParseJSON = roleSpecController.ParseJSON
	container.RegistController("roleSpec", roleSpecController)
}
