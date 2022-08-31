package rest

import (
	"template/internal/api/rest/internal"

	"github.com/gin-gonic/gin"
)

// Hello godoc
// @Summary 问候
// @Tags Hello
// @Description get greet by name
// @Accept json
// @Produce json
// @Param name path string true "昵称" default(libz)
// @Param code query string true "区服编号" default(1001)
// @Param Content-Type header string true "数据格式" default(application/json)
// @Success 200 {object} internal.Response{data=internal.HelloRsp} "响应体"
// @Router /v1/hello/{name} [get]
func (c *restHandler) Hello(ctx *gin.Context) {
	data, err := c.useCase.Hello(ctx.Request.Context(), ctx.Param("name"))
	if err != nil {
		c.ResponseWithDesc(ctx, internal.CodePlayerInfo, err.Error())
		return
	}

	rsp := &internal.HelloRsp{Greet: data}
	c.ResponseWithData(ctx, rsp)
}
