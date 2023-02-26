package controller

import (
	"github.com/curltech/go-colla-core/logger"
	"github.com/curltech/go-colla-core/service"
	"github.com/curltech/go-colla-core/util/message"
	"github.com/curltech/go-colla-core/util/reflect"
	"github.com/kataras/iris/v12"
	"strings"
)

type QueryParas struct {
	CondiBean  interface{}
	Data       []interface{}
	Conditions []interface{}
	Columns    []string
}

type BaseController struct {
	BaseService service.BaseService
	ParseJSON   func(json []byte) (interface{}, error)
}

var baseController *BaseController = &BaseController{BaseService: service.GetOrmBaseService()}

func GetBaseController() *BaseController {
	return baseController
}

type PageParam struct {
	From      int         `json:"from,omitempty"`
	Limit     int         `json:"limit,omitempty"`
	Count     int64       `json:"count,omitempty"`
	Orderby   string      `json:"orderby,omitempty"`
	CondiBean interface{} `json:"condiBean,omitempty"`
}

/**
返回读取json数据转换成相应的输入数据，1.实体数组;2.实体;3.分页参数，含有条件实体;4.映射
后面三项返回的参数放入数组中
*/
func (this *BaseController) ReadJSON(ctx iris.Context) ([]interface{}, error) {
	// 先按照数组结构解析
	json, err := ctx.GetBody()
	if err != nil {
		return nil, err
	}
	logger.Sugar.Infof(string(json))
	rowsSlicePtr := make([]interface{}, 0)
	entities, err := this.ParseJSON(json)
	if err != nil { //解析有错，按照单个实体解析
		logger.Sugar.Errorf("ReadJSON entities exception:%v", err)
		entity, err := this.BaseService.NewEntity(nil)
		if err != nil {
			logger.Sugar.Errorf("NewEntity exception:%v", err)

			return nil, err
		}
		if strings.Contains(string(json), "limit") && strings.Contains(string(json), "condiBean") {
			pageParam := &PageParam{CondiBean: entity}
			err = message.Unmarshal(json, pageParam)
			if err != nil { //都不能解析，出错
				logger.Sugar.Errorf("ReadJSON pageParam exception:%v", err)
			} else {
				rowsSlicePtr = append(rowsSlicePtr, pageParam)
			}
		}
		if len(rowsSlicePtr) == 0 {
			err = message.Unmarshal(json, entity)
			if err != nil { //都不能解析，出错
				logger.Sugar.Errorf("ReadJSON entity exception:%v", err)
				condiBean := make(map[string]interface{})
				err = message.Unmarshal(json, &condiBean)
				if err != nil { //都不能解析，出错
					logger.Sugar.Errorf("ReadJSON condiBean exception:%v", err)

					return nil, err
				} else {
					rowsSlicePtr = append(rowsSlicePtr, condiBean)
				}
			} else {
				//能够解析，放入数组，统一以数组形式返回
				rowsSlicePtr = append(rowsSlicePtr, entity)
			}
		}
	} else {
		rowsSlicePtr = reflect.ToArray(entities)
	}
	// 返回数据数组
	return rowsSlicePtr, nil
}

// Get retrieve one record from database, bean's non-empty fields
// will be as conditions
func (this *BaseController) Get(ctx iris.Context) {
	//cond := make(map[string]interface{},0)
	orderby := ctx.URLParam("orderby")
	condiBean, _ := this.ReadJSON(ctx)
	result, _ := this.BaseService.Get(condiBean[0], false, orderby, "")
	if result {
		ctx.JSON(condiBean)
	} else {
		ctx.StopWithJSON(iris.StatusInternalServerError, "GetFail")
	}
}

// Find retrieve records from table, condiBeans's non-empty fields
// are conditions. beans could be []Struct, []*Struct, map[int64]Struct
// map[int64]*Struct everyone := make([]Userinfo, 0)
// err := engine.Find(&everyone)
func (this *BaseController) Find(ctx iris.Context) {
	data, err := this.BaseService.NewEntities(nil)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())

		return
	}
	condiBeans, err := this.ReadJSON(ctx)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
	var pageParam *PageParam
	var condiBean interface{}
	var from int
	var limit int
	var count int64
	var orderby string
	params := condiBeans[0]
	pageParam, ok := params.(*PageParam)
	if ok {
		condiBean = pageParam.CondiBean
		from = pageParam.From
		limit = pageParam.Limit
		count = pageParam.Count
		orderby = pageParam.Orderby
	} else {
		condiBean = params
	}
	if pageParam != nil && count == 0 {
		count, _ = this.BaseService.Count(condiBean, "")
	}
	if pageParam == nil || count > 0 {
		err = this.BaseService.Find(data, condiBean, orderby, from, limit, "")
		if err != nil {
			ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
		}
	}
	result := make(map[string]interface{})
	if pageParam != nil {
		result["count"] = count
	}

	result["data"] = data
	ctx.JSON(result)
}

// insert model data to database
func (this *BaseController) Insert(ctx iris.Context) {
	data, err := this.ReadJSON(ctx)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())

		return
	}
	affected, _ := this.BaseService.Insert(data...)
	if affected > 0 {
		ctx.JSON(&data)
	} else {
		ctx.StopWithJSON(iris.StatusInternalServerError, "InsertFail")
	}
}

// update model to database.
// cols set the columns those want to update.
func (this *BaseController) Update(ctx iris.Context) {
	//cond := make(map[string]interface{})
	data, err := this.ReadJSON(ctx)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())

		return
	}
	affected, _ := this.BaseService.Update(data, nil, "")
	if affected > 0 {
		ctx.JSON(data)
	} else {
		ctx.StopWithJSON(iris.StatusInternalServerError, "UpdateFail")
	}
}

func (this *BaseController) Upsert(ctx iris.Context) {
	//cond := make(map[string]interface{})
	data, err := this.ReadJSON(ctx)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())

		return
	}
	affected, _ := this.BaseService.Upsert(data...)
	if affected > 0 {
		ctx.JSON(data)
	} else {
		ctx.StopWithJSON(iris.StatusInternalServerError, "UpsertFail")
	}
}

// delete model in database
// Delete records, bean's non-empty fields are conditions
func (this *BaseController) Delete(ctx iris.Context) {
	//cond := make(map[string]interface{})
	data, err := this.ReadJSON(ctx)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())

		return
	}
	affected, _ := this.BaseService.Delete(data, "")
	if affected > 0 {
		ctx.JSON(data)
	} else {
		ctx.StopWithJSON(iris.StatusInternalServerError, "DeleteFail")
	}
}

func (this *BaseController) Save(ctx iris.Context) {
	//cond := make(map[string]interface{})
	data, err := this.ReadJSON(ctx)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())

		return
	}
	affected, _ := this.BaseService.Save(data...)
	if affected > 0 {
		ctx.JSON(data)
	} else {
		ctx.StopWithJSON(iris.StatusInternalServerError, "SaveFail")
	}
}

func init() {

}
