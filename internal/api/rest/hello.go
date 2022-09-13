package rest

import (
	"template/internal/api/rest/internal"

	"github.com/gin-gonic/gin"
)

// Hello godoc
// @Summary 问候
// @Tags Hello
// @Description 这里写一大段描述
// @Description 还支持多行
// @Accept json
// @Produce json
// @Param 	name 			path 	string 		true 	"昵称"			default(libz)
// @Param 	code 			query 	string 		true 	"区服编号"		default(1001)
// @Param 	Content-Type 	header 	string 		true 	"数据格式" default(application/json)
// @Param	body			body	object{name=string,age=int}		true	"测试数据" default({"name":"libz", "age":10})
// @Success 200				{object}	object{code=int,message=string,data=internal.HelloRsp} "响应体"
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
