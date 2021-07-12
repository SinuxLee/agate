package rest

import (
	"net/http"
	"template/internal/api/rest/internal"
	"template/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var _ Handler = (*restHandler)(nil)

type Handler interface {
	Hello(*gin.Context)
}

func NewRestHandler(uc service.UseCase) Handler {
	return &restHandler{
		useCase: uc,
	}
}

type restHandler struct {
	useCase service.UseCase
}

// ResponseWithData ...
func (c *restHandler) ResponseWithData(ctx *gin.Context, data interface{}) {
	c.innerResponse(ctx, &internal.Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

// ResponseWithCode ...
func (c *restHandler) ResponseWithCode(ctx *gin.Context, code int) {
	resp := &internal.Response{Code: code}
	desc, ok := internal.CodeText[code]
	if ok {
		resp.Msg = desc
	} else {
		resp.Msg = "unknown error"
	}

	c.innerResponse(ctx, resp)
}

// ResponseWithDesc ...
func (c *restHandler) ResponseWithDesc(ctx *gin.Context, code int, desc string) {
	c.innerResponse(ctx, &internal.Response{
		Code: code,
		Msg:  desc,
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
	raw, _ := ctx.GetRawData()
	log.Error().Str("path", ctx.Request.URL.Path).
		Str("query", ctx.Request.URL.RawQuery).
		Str("request", string(raw)).
		Interface("response", resp).
		Msg("bad response")
}

func healthCheck(engine *gin.Engine) {
	engine.GET("/healthz", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "It is OK\n")
	})
}

func RegisterHandler(engine *gin.Engine, ctrl Handler) {
	healthCheck(engine)

	group1 := engine.Group("/svr/v1")
	group1.GET("hello", ctrl.Hello)
}
