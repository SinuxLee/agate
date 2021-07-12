package rest

import (
	"template/internal/api/rest/internal"

	"github.com/asim/go-micro/v3/logger"
	"github.com/gin-gonic/gin"
)

func (c *restHandler) Hello(ctx *gin.Context) {
	data, err := c.useCase.Hello(ctx.Request.Context(), "libz")
	if err != nil {
		c.ResponseWithDesc(ctx, internal.CodePlayerInfo, err.Error())
		return
	}

	rsp := &internal.HelloRsp{Greet: data}
	logger.Infof("%+v", *rsp)

	c.ResponseWithData(ctx, rsp)
}
