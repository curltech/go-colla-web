package controller

import (
	"errors"
	"github.com/casbin/casbin/v2"
	"github.com/curltech/go-colla-biz/rbac"
	"github.com/curltech/go-colla-biz/rbac/entity"
	"github.com/curltech/go-colla-biz/rbac/service"
	"github.com/curltech/go-colla-core/cache"
	"github.com/curltech/go-colla-core/config"
	"github.com/curltech/go-colla-core/logger"
	cas "github.com/iris-contrib/middleware/casbin"
	"github.com/kataras/iris/v12"
	"net/http"
	"path/filepath"
)

var CasbinMemCache = cache.NewMemCache("casbin", 60, 10)

var casbinMiddleware *cas.Casbin
var enforcer *casbin.Enforcer

/**
设置casbin拦截器控制权限
*/
func Set(app *iris.Application) {
	adapter := rbac.GetAdapter()
	var err error
	enforcer, err = casbin.NewEnforcer(config.RbacParams.Model, adapter)
	if err != nil {
		logger.Sugar.Error(err.Error())
	}
	casbinMiddleware = cas.New(enforcer)
	app.Use(ServeHTTP)
}

func ServeHTTP(ctx iris.Context) {
	if config.AppParams.EnableSession {
		user := CurrentUser(ctx)
		if user == nil {
			logger.Sugar.Error("NoUser")
			ctx.StopWithJSON(http.StatusUnauthorized, "NoUser") // Status Forbidden

			return
		} else {
			_, ok := checkNone(ctx)
			if ok {
				ctx.Next()
				return
			}
			err := Check(ctx, user)
			if err != nil {
				logger.Sugar.Error(err.Error())
				ctx.StopWithJSON(http.StatusForbidden, err.Error()) // Status Forbidden

				return
			}
		}
	}
	ctx.Next()
}

func Check(ctx iris.Context, currentUser *entity.User) error {
	path, ok := checkNone(ctx)
	if ok {
		return nil
	}
	if config.RbacParams.ValidResource == true {
		resource := entity.Resource{}
		resource.Status = entity.UserStatus_Enabled
		resource.Path = path
		svc := service.GetResourceService()
		exist, _ := svc.Get(&resource, false, "", "")
		if exist == false {
			return nil
		}
	}

	method := ctx.Request().Method
	ok, err := enforcer.Enforce(currentUser.UserId, path, method)
	if err != nil {
		logger.Sugar.Error(err.Error())
		return err
	}
	if !ok {
		logger.Sugar.Error("NoAuth")
		return errors.New("NoAuth")
	}

	return nil
}

func checkNone(ctx iris.Context) (string, bool) {
	addr := ctx.Request().RemoteAddr
	if len(config.RbacParams.NoneAddress) > 0 {
		for _, pattern := range config.RbacParams.NoneAddress {
			matched, err := filepath.Match(pattern, addr)
			if err == nil && matched == true {
				logger.Sugar.Infof("Address %s match pattern %s", addr, pattern)
				return "", true
			}
		}
	}
	path := ctx.Request().URL.Path
	if len(config.RbacParams.NonePath) > 0 {
		for _, pattern := range config.RbacParams.NonePath {
			matched, err := filepath.Match(pattern, path)
			if err == nil && matched == true {
				logger.Sugar.Infof("Path %s match pattern %s", path, pattern)
				return path, true
			}
		}
	}
	return "", false
}
