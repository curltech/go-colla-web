package controller

import (
	"errors"
	"fmt"
	"github.com/curltech/go-colla-core/config"
	"github.com/curltech/go-colla-core/container"
	"github.com/curltech/go-colla-core/logger"
	"github.com/curltech/go-colla-core/util/debug"
	"github.com/curltech/go-colla-core/util/reflect"
	"github.com/go-playground/validator/v10"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"mime/multipart"
	"strings"
)

var TemplateParams = make(map[string]interface{})

func RegistTemplateParam(key string, param interface{}) error {
	_, ok := TemplateParams[key]
	if !ok {
		TemplateParams[key] = param

		return nil
	}

	return errors.New("Exist")
}

func init() {
	RegistTemplateParam("Host", config.ServerParams.Name)
}

func HTMLController(ctx iris.Context) {
	name := ctx.Params().GetString("name")
	name = strings.ReplaceAll(name, ".html", "")
	fn := debug.TraceDebug("render view:" + name)
	defer fn()
	ctx.View(name, TemplateParams)
}

func MainController(ctx iris.Context) {
	serviceName := ctx.Params().Get("serviceName")
	methodName := ctx.Params().Get("methodName")
	logger.Sugar.Infof("MainController call %v.%v", serviceName, methodName)
	args := make([]interface{}, 1)
	args[0] = ctx
	msg := fmt.Sprintf("call servicename:%v methodName:%v", serviceName, methodName)
	fn := debug.TraceDebug(msg)
	defer fn()
	controller := container.GetController(serviceName)
	if controller == nil {
		logger.Sugar.Error("NoController:%s", serviceName)
		ctx.StopWithJSON(iris.StatusInternalServerError, "NoController")

		return
	}
	_, err := reflect.Call(controller, methodName, args)
	if err != nil {
		logger.Sugar.Error(err.Error())
	}
}

type UploadParam struct {
	ServiceName string
}

const postMaxSize = 256 * iris.MB

func UploadController(ctx iris.Context) {
	ctx.SetMaxRequestBodySize(postMaxSize)
	var serviceName = ctx.PostValue("serviceName")
	var methodName = ctx.PostValue("methodName")
	if serviceName == "" || methodName == "" {
		logger.Sugar.Error("NoService:%s", serviceName)
		ctx.StopWithJSON(iris.StatusInternalServerError, "NoService")

		return
	}
	logger.Sugar.Infof("UploadController call %v.%v", serviceName, methodName)

	err := ctx.Request().ParseMultipartForm(postMaxSize)
	if err != nil {
		logger.Sugar.Error(err.Error())
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())

		return
	}
	form := ctx.Request().MultipartForm
	if form == nil {
		logger.Sugar.Error("BlankForm")
		ctx.StopWithJSON(iris.StatusInternalServerError, "BlankForm")

		return
	}
	if form.File == nil {
		logger.Sugar.Error("NilFormFile")
		ctx.StopWithJSON(iris.StatusInternalServerError, "NilFormFile")

		return
	}
	var files = make([]multipart.File, 0)
	for _, heads := range form.File {
		for _, head := range heads {
			logger.Sugar.Infof("file:%v", head.Filename)
			file, err := head.Open()
			if err != nil {
				logger.Sugar.Error(err.Error())
				ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())

				return
			}
			files = append(files, file)
			/**
			下面是读取数据到[]byte
			*/
			//buf, err := ioutil.ReadAll(file)
			//if err != nil {
			//	ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())
			//
			//	return
			//}
			defer file.Close()
		}
	}

	args := make([]interface{}, 1)
	args[0] = files
	msg := fmt.Sprintf("call servicename:%v methodName:%v", serviceName, methodName)
	fn := debug.TraceDebug(msg)
	defer fn()
	svc := container.GetService(serviceName)
	if svc == nil {
		logger.Sugar.Errorf("NoService:%s", serviceName)
		ctx.StopWithJSON(iris.StatusInternalServerError, "NoService")

		return
	}
	result, err := reflect.Call(svc, methodName, args)
	if err != nil {
		logger.Sugar.Error(err.Error())
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())

		return
	} else {
		ctx.JSON(result)
	}
}

func DownloadController(ctx iris.Context) {
	params := make(map[string]interface{}, 0)
	ctx.ReadJSON(&params)

	serviceName := params["serviceName"]
	methodName := params["methodName"]
	destName := params["destName"]

	logger.Sugar.Infof("DownloadController call %v.%v", serviceName, methodName)
	ctx.ContentType("")
	ctx.ResponseWriter().Header().Set(context.ContentDispositionHeaderKey, "attachment;filename="+destName.(string))
	args := make([]interface{}, 2)
	args[0] = params["condiBean"]
	msg := fmt.Sprintf("call servicename:%v methodName:%v", serviceName, methodName)
	fn := debug.TraceDebug(msg)
	defer fn()
	svc := container.GetService(serviceName.(string))
	if svc == nil {
		panic("NoService")
	}
	result, err := reflect.Call(svc, methodName.(string), args)
	if err != nil {
		logger.Sugar.Error(err.Error())
		ctx.JSON(err.Error())
	} else {
		ctx.ResponseWriter().Write(result[0].([]byte))
		ctx.ResponseWriter().Flush()
	}
}

type validationError struct {
	ActualTag string `json:"tag"`
	Namespace string `json:"namespace"`
	Kind      string `json:"kind"`
	Type      string `json:"type"`
	Value     string `json:"value"`
	Param     string `json:"param"`
}

func wrapValidationErrors(errs validator.ValidationErrors) []validationError {
	validationErrors := make([]validationError, 0, len(errs))
	for _, validationErr := range errs {
		validationErrors = append(validationErrors, validationError{
			ActualTag: validationErr.ActualTag(),
			Namespace: validationErr.Namespace(),
			Kind:      validationErr.Kind().String(),
			Type:      validationErr.Type().String(),
			Value:     fmt.Sprintf("%v", validationErr.Value()),
			Param:     validationErr.Param(),
		})
	}

	return validationErrors
}
