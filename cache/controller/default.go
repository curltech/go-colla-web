package controller

import (
	"github.com/curltech/go-colla-core/cache"
	"github.com/curltech/go-colla-core/container"
	"github.com/kataras/iris/v12"
	cache2 "github.com/patrickmn/go-cache"
)

type CacheInfo struct {
	Name  string
	Items map[string]cache2.Item
}

type CacheController struct {
}

var cacheController *CacheController

func (this *CacheController) GetCaches(ctx iris.Context) {
	cacheInfos := make([]*CacheInfo, 0)
	for name, c := range cache.MemCaches {
		cacheInfo := CacheInfo{Name: name, Items: c.Items()}
		cacheInfos = append(cacheInfos, &cacheInfo)
	}
	ctx.JSON(cacheInfos)
}

func (this *CacheController) Flush(ctx iris.Context) {
	condiBean := make(map[string]interface{})
	err := ctx.ReadJSON(&condiBean)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())

		return
	}
	v, ok := condiBean["name"]
	if ok {
		name := v.(string)
		var key string
		v, ok = condiBean["key"]
		if ok {
			key = v.(string)
		}
		if key == "" {
			cache.MemCaches[name].Flush()
		} else {
			cache.MemCaches[name].Delete(key)
		}
	} else {
		ctx.StopWithJSON(iris.StatusInternalServerError, "NoName")
	}
}

func init() {
	container.RegistController("cache", cacheController)
}
