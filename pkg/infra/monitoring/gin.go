package monitoring

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/zsais/go-gin-prometheus"
)

func GinHandler() gin.HandlerFunc {
	ginProm := ginprometheus.NewPrometheus(defaultConf.ServerName)

	paramSet := map[string]string{
		"name": ":name",
	}

	ginProm.ReqCntURLLabelMappingFn = func(c *gin.Context) string {
		url := c.Request.URL.Path
		for _, param := range c.Params {
			if value, exist := paramSet[param.Key]; exist {
				url = strings.Replace(url, param.Value, value, 1)
				break
			}
		}
		return url
	}
	return ginProm.HandlerFunc()
}
