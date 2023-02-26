package controller

import (
	"github.com/curltech/go-colla-biz/controller"
	"github.com/curltech/go-colla-core/base/entity"
	"github.com/curltech/go-colla-core/base/service"
	"github.com/curltech/go-colla-core/config"
	"github.com/curltech/go-colla-core/container"
	entity2 "github.com/curltech/go-colla-core/entity"
	"github.com/curltech/go-colla-core/logger"
	"github.com/curltech/go-colla-core/util/message"
	"github.com/kataras/iris/v12"
	"time"
)

/**
控制层代码需要做数据转换，调用服务层的代码，由于数据转换的结构不一致，因此每个实体（外部rest方式访问）的控制层都需要写一遍
*/
type SessionInstanceController struct {
	controller.BaseController
}

var sessionInstanceController *SessionInstanceController

func (this *SessionInstanceController) ParseJSON(json []byte) (interface{}, error) {
	var entities = make([]*entity.SessionInstance, 0)
	err := message.Unmarshal(json, &entities)

	return &entities, err
}

/**
注册bean管理器，注册序列
*/
func init() {
	sessionInstanceController = &SessionInstanceController{
		BaseController: controller.BaseController{
			BaseService: service.GetSessionInstanceService(),
		},
	}
	sessionInstanceController.BaseController.ParseJSON = sessionInstanceController.ParseJSON
	container.RegistController("sessionInstance", sessionInstanceController)
}

/**
session处理，获取当前会话，不存在创建，否则获取会话ID，存入数据库记录
放在iris全局最早的中间件
*/
func SessionController(ctx iris.Context) {
	session := controller.GetSession().Start(ctx)
	if config.AppParams.SessionLog == false {
		ctx.Next()
		return
	}
	sessionId := session.ID()
	now := time.Now()
	sessionInstance := entity.SessionInstance{
		Host:      ctx.Host(),
		SessionId: sessionId,
		//Locale:    ctx.GetLocale().Language(),
		Url:            ctx.FullRequestURI(),
		IsMobile:       ctx.IsMobile(),
		IsSSL:          ctx.IsSSL(),
		LastAccessTime: &now,
	}
	if session.IsNew() {
		logger.Sugar.Infof("new session:%v", sessionId)
		sessionInstance.Status = entity2.EntityState_New
	} else {
		logger.Sugar.Infof("exist session access:%v", sessionId)
		sessionInstance.Status = entity2.EntityState_Modified
	}
	sessionInstanceService := service.GetSessionInstanceService()
	go sessionInstanceService.Insert(&sessionInstance)
	ctx.Next()
}
