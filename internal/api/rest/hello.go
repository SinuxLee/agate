package rest

import (
	"template/internal/api/rest/internal"

	"github.com/asim/go-micro/v3/logger"
	"github.com/gin-gonic/gin"
)

// Hello godoc
// @Summary 问候
// @Tags Hello
// @Description get greet by name
// @Accept json
// @Produce json
// @Param name path string true "libz"
// @Param Content-Type header string true "application/json" default(application/json)
// @Success 200 {object} internal.Response{data=internal.HelloRsp} "响应体"
// @Router /v1/hello/{name} [get]
func (c *restHandler) Hello(ctx *gin.Context) {
	data, err := c.useCase.Hello(ctx.Request.Context(), ctx.Param("name"))
	if err != nil {
		c.ResponseWithDesc(ctx, internal.CodePlayerInfo, err.Error())
		return
	}

	rsp := &internal.HelloRsp{Greet: data}
	logger.Infof("%+v", *rsp)

	c.ResponseWithData(ctx, rsp)
}
