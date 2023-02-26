package controller

import (
	"github.com/curltech/go-colla-biz/actual/businessentity"
	"github.com/curltech/go-colla-biz/actual/entity"
	entity2 "github.com/curltech/go-colla-biz/spec/entity"
	"github.com/curltech/go-colla-core/container"
	baseerror "github.com/curltech/go-colla-core/error"
	"github.com/kataras/iris/v12"
)

type BusinessEntityController struct {
}

var businessEntityController *BusinessEntityController

func (this *BusinessEntityController) Create(ctx iris.Context) {
	condiBean := entity.Role{}
	err := ctx.ReadJSON(&condiBean)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
	be, err := businessentity.GetBusinessEntityService().Create(condiBean.SchemaName, condiBean.SpecId, condiBean.EffectiveDate)
	if be != nil {
		ctx.JSON(be.GetActual(true))
	} else {
		ctx.StopWithJSON(iris.StatusInternalServerError, baseerror.Error_NotFound)
	}
}

func (this *BusinessEntityController) Save(ctx iris.Context) {
	condiBean := entity.Role{}
	err := ctx.ReadJSON(&condiBean)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
	be, err := businessentity.GetBusinessEntityService().Save(condiBean.SchemaName, condiBean.Id)
	if be != nil {
		ctx.JSON(be.GetActual(true))
	} else {
		ctx.StopWithJSON(iris.StatusInternalServerError, baseerror.Error_NotFound)
	}
}

func (this *BusinessEntityController) AddRole(ctx iris.Context) {
	condiBean := entity.Role{}
	err := ctx.ReadJSON(&condiBean)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
	be, _ := businessentity.GetBusinessEntityService().AddRole(condiBean.SchemaName, condiBean.Id, condiBean.ParentId, condiBean.Kind, nil)
	if be != nil {
		ctx.JSON(be.GetActual(true))
	} else {
		ctx.StopWithJSON(iris.StatusInternalServerError, baseerror.Error_NotFound)
	}
}

func (this *BusinessEntityController) LoadRole(ctx iris.Context) {
	condiBean := entity.Role{}
	err := ctx.ReadJSON(&condiBean)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
	be, _ := businessentity.GetBusinessEntityService().LoadRole(condiBean.SchemaName, condiBean.Id, condiBean.ParentId, condiBean.Kind)
	if be != nil {
		ctx.JSON(be.GetActual(true))
	} else {
		ctx.StopWithJSON(iris.StatusInternalServerError, baseerror.Error_NotFound)
	}
}

func (this *BusinessEntityController) Delete(ctx iris.Context) {
	condiBean := entity.Role{}
	err := ctx.ReadJSON(&condiBean)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
	affected, err := businessentity.GetBusinessEntityService().Delete(condiBean.SchemaName, condiBean.Id)
	if err == nil {
		ctx.JSON(affected)
	} else {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
}

func (this *BusinessEntityController) RemoveRole(ctx iris.Context) {
	condiBean := entity.Role{}
	err := ctx.ReadJSON(&condiBean)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
	be, err := businessentity.GetBusinessEntityService().RemoveRole(condiBean.SchemaName, condiBean.TopId, condiBean.Id, condiBean.Path)
	if err == nil {
		ctx.JSON(be.GetActual(true))
	} else {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
}

func (this *BusinessEntityController) Get(ctx iris.Context) {
	condiBean := entity.Role{}
	err := ctx.ReadJSON(&condiBean)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
	be := businessentity.GetBusinessEntityService().Get(condiBean.SchemaName, condiBean.Id)
	if be != nil {
		ctx.JSON(be.GetActual(false))
	} else {
		ctx.StopWithJSON(iris.StatusInternalServerError, baseerror.Error_NotFound)
	}
}

func (this *BusinessEntityController) Update(ctx iris.Context) {
	actual := make(map[string]interface{}, 0)
	err := ctx.ReadJSON(&actual)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
	be, _ := businessentity.GetBusinessEntityService().Update(actual)
	if be != nil {
		ctx.JSON(true)
	} else {
		ctx.StopWithJSON(iris.StatusInternalServerError, baseerror.Error_NotFound)
	}
}

func (this *BusinessEntityController) SetValue(ctx iris.Context) {
	actual := make(map[string]interface{}, 0)
	err := ctx.ReadJSON(&actual)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
	be, err := businessentity.GetBusinessEntityService().SetValue(actual)
	if be != nil {
		if err == nil {
			ctx.JSON(be.GetActual(false))
		} else {
			ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
		}
	} else {
		ctx.StopWithJSON(iris.StatusInternalServerError, baseerror.Error_NotFound)
	}
}

func (this *BusinessEntityController) Load(ctx iris.Context) {
	condiBean := entity.Role{}
	err := ctx.ReadJSON(&condiBean)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
	be := businessentity.GetBusinessEntityService().Load(condiBean.SchemaName, condiBean.Id)
	if be != nil {
		ctx.JSON(be.GetActual(false))
	} else {
		ctx.StopWithJSON(iris.StatusInternalServerError, baseerror.Error_NotFound)
	}
}

func (this *BusinessEntityController) Version(ctx iris.Context) {
	condiBean := entity.Role{}
	err := ctx.ReadJSON(&condiBean)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
	be, err := businessentity.GetBusinessEntityService().Version(condiBean.SchemaName, condiBean.Id)
	if be != nil {
		ctx.JSON(be.GetActual(false))
	} else {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
}

func (this *BusinessEntityController) Find(ctx iris.Context) {
	condiBean := entity.Role{}
	err := ctx.ReadJSON(&condiBean)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
	v, specType, ok := businessentity.GetBusinessEntityService().Find(condiBean.SchemaName, condiBean.Id, condiBean.Path, false)
	if ok {
		if specType == entity2.SpecType_Role {
			be := v.(*businessentity.BusinessEntity)
			ctx.JSON(be.GetActual(false))
		} else {
			ctx.JSON(v)
		}
	} else {
		ctx.StopWithJSON(iris.StatusInternalServerError, baseerror.Error_NotExist)
	}
}

func (this *BusinessEntityController) GetActual(ctx iris.Context) {
	condiBean := entity.Role{}
	err := ctx.ReadJSON(&condiBean)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
	actual := businessentity.GetBusinessEntityService().GetActual(condiBean.SchemaName, condiBean.Id, false)
	if actual != nil {
		ctx.JSON(actual)
	} else {
		ctx.StopWithJSON(iris.StatusInternalServerError, baseerror.Error_NotFound)
	}
}

func init() {
	container.RegistController("businessEntity", businessEntityController)
}
