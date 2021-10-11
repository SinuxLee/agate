package rest

import (
	"fmt"
	"net/http"

	"template/internal/api/rest/docs"
	"template/internal/api/rest/internal"
	"template/internal/service"
	"template/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

var _ Handler = (*restHandler)(nil)

type Handler interface {
	RegisterHandler(engine *gin.Engine)
}

func NewRestHandler(uc service.UseCase, swaggerAddr string) Handler {
	return &restHandler{
		useCase:     uc,
		swaggerHost: swaggerAddr,
	}
}

// swagger 项目描述
// @title template
// @version 1.0
// @description 项目结构概要描述
// @termsOfService http://swagger.io/terms/

// @tag.name Hello
// @tag.description 各种问候

// @contact.name sinuxlee
// @contact.url http://www.swagger.io/support
// @contact.email sinuxlee@qq.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @schemes http https
// @host localhost:8086
// @BasePath /svr
// @query.collection.format multi

// @securityDefinitions.basic BasicAuth

// @securityDefinitions.apikey TokenAuth
// @in header
// @name Authorization

// @x-extension-openapi {"example": "value on a json format"}

type restHandler struct {
	useCase     service.UseCase
	swaggerHost string
}

// ResponseWithData ...
func (c *restHandler) ResponseWithData(ctx *gin.Context, data interface{}) {
	c.innerResponse(ctx, &internal.Response{
		Code:    0,
		Message: "OK",
		Data:    data,
	})
}

// ResponseWithCode ...
func (c *restHandler) ResponseWithCode(ctx *gin.Context, code int) {
	resp := &internal.Response{Code: code}
	desc, ok := internal.CodeText[code]
	if ok {
		resp.Message = desc
	} else {
		resp.Message = "unknown error"
	}

	c.innerResponse(ctx, resp)
}

// ResponseWithDesc ...
func (c *restHandler) ResponseWithDesc(ctx *gin.Context, code int, desc string) {
	c.innerResponse(ctx, &internal.Response{
		Code:    code,
		Message: desc,
	})
}

func (c *restHandler) innerResponse(ctx *gin.Context, resp *internal.Response) {
	ctx.Header("X-Robot-Index", ctx.GetHeader("X-Robot-Index"))
	ctx.JSON(http.StatusOK, resp)
	if resp.Code != internal.CodeSuccess {
		c.ErrorLog(ctx, resp)
	}
}

func (c *restHandler) ErrorLog(ctx *gin.Context, resp *internal.Response) {
	log.Error().Str("path", ctx.Request.URL.Path).
		Str("query", ctx.Request.URL.RawQuery).
		Interface("response", resp).
		Msg("bad response")
}

func (c *restHandler) swaggerDocs(engine *gin.Engine) {
	docs.SwaggerInfo.Host = c.swaggerHost
	url := ginSwagger.URL(fmt.Sprintf("http://%v/swagger/doc.json", c.swaggerHost))
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
}

func (c *restHandler) healthCheck(engine *gin.Engine) {
	engine.GET("/healthz", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "It is OK\n")
	})
}

func (c *restHandler) RegisterHandler(engine *gin.Engine) {
	c.healthCheck(engine)
	c.swaggerDocs(engine)

	group1 := engine.Group("/svr/v1")
	group1.Use(middleware.Logger())
	group1.GET("hello/:name", c.Hello)
	group1.POST("hello/:name", c.Hello)
}
