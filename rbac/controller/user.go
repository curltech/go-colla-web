package controller

import (
	"github.com/curltech/go-colla-biz/controller"
	"github.com/curltech/go-colla-biz/rbac/entity"
	service2 "github.com/curltech/go-colla-biz/rbac/service"
	"github.com/curltech/go-colla-core/cache"
	"github.com/curltech/go-colla-core/config"
	"github.com/curltech/go-colla-core/container"
	"github.com/curltech/go-colla-core/logger"
	"github.com/curltech/go-colla-core/util/json"
	"github.com/curltech/go-colla-core/util/message"
	"github.com/kataras/iris/v12"
)

var MemCache = cache.NewMemCache("sessionUser", 60, 10)

/**
控制层代码需要做数据转换，调用服务层的代码，由于数据转换的结构不一致，因此每个实体（外部rest方式访问）的控制层都需要写一遍
*/
type UserController struct {
	controller.BaseController
}

var userController *UserController

func GetUserController() *UserController {
	return userController
}

func (this *UserController) ParseJSON(json []byte) (interface{}, error) {
	var entities = make([]*entity.User, 0)
	err := message.Unmarshal(json, &entities)

	return &entities, err
}

func (this *UserController) getSessionCacheKey(ctx iris.Context) string {
	sess := controller.GetSession().Start(ctx)
	sessionId := sess.ID()

	return "session:" + sessionId
}

func (this *UserController) Regist(ctx iris.Context) {
	user := &entity.User{}
	err := ctx.ReadJSON(user)
	if err != nil {
		logger.Sugar.Error(err.Error())
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
	}
	service := this.BaseService.(*service2.UserService)
	user, err = service.Regist(user)
	if err != nil {
		logger.Sugar.Error(err.Error())
		ctx.StopWithJSON(iris.StatusOK, err.Error())
	} else {
		user.Password = ""
		user.PlainPassword = ""
		user.ConfirmPassword = ""
		ctx.JSON(user)
	}
}

func (this *UserController) GetCurrentUser(ctx iris.Context) {
	user := this.CurrentUser(ctx)
	if user != nil {
		ctx.JSON(user)
	} else {
		logger.Sugar.Error("NoUser")
		ctx.StopWithJSON(iris.StatusInternalServerError, "NoUser")
	}
}

func (this *UserController) GetCurrentUserName(ctx iris.Context) string {
	if config.AppParams.EnableSession {
		key := this.getSessionCacheKey(ctx)
		var userName string
		v, ok := MemCache.Get(key)
		if ok {
			userName = v.(string)
			return userName
		}
	}

	return ""
}

func (this *UserController) CurrentUser(ctx iris.Context) *entity.User {
	if config.AppParams.EnableSession {
		var userName = this.GetCurrentUserName(ctx)
		if userName != "" {
			service := this.BaseService.(*service2.UserService)

			return service.GetUser(userName)
		}
	}

	return nil
}

func (this *UserController) Logout(ctx iris.Context) {
	if config.AppParams.EnableSession {
		key := this.getSessionCacheKey(ctx)
		var user = this.CurrentUser(ctx)
		if user != nil {
			service := this.BaseService.(*service2.UserService)
			service.Logout(user.UserName)

			MemCache.Delete(key)
		}
	}
	ctx.Logout()
}

func (this *UserController) Login(ctx iris.Context) {
	this.Logout(ctx)
	params := make(map[string]string, 0)
	err := ctx.ReadJSON(&params)
	service := this.BaseService.(*service2.UserService)
	user, err := service.Login(params[config.RbacParams.Credential], params[config.RbacParams.Password])
	if err != nil {
		logger.Sugar.Error(err.Error())
		ctx.StopWithJSON(iris.StatusOK, "NoUser")
	} else {
		result := make(map[string]interface{})
		if config.AppParams.EnableSession {
			key := this.getSessionCacheKey(ctx)
			MemCache.SetDefault(key, user.UserName)
		}
		if config.AppParams.EnableJwt {
			token := GenerateToken(ctx, user)
			if token != nil {
				result["token"] = string(token)
			} else {
				logger.Sugar.Error("NilToken")
				ctx.StopWithJSON(iris.StatusOK, "NilToken")
			}
		}
		b, err := json.Marshal(user)
		if err == nil {
			u := entity.User{}
			err = json.Unmarshal(b, &u)
			if err == nil {
				u.Password = ""
				u.PlainPassword = ""
				u.ConfirmPassword = ""
				result["user"] = u
				ctx.JSON(result)
			} else {
				logger.Sugar.Error(err.Error())
				ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
			}
		} else {
			logger.Sugar.Error(err.Error())
			ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
		}
	}
}

/**
注册bean管理器，注册序列
*/
func init() {
	userController = &UserController{
		BaseController: controller.BaseController{
			BaseService: service2.GetUserService(),
		},
	}
	userController.BaseController.ParseJSON = userController.ParseJSON
	container.RegistController("user", userController)
}
