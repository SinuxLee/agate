package rest

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"template/internal/api/rest/docs"
	"template/internal/api/rest/internal"
	"template/internal/entity"
	"template/internal/service"
	"template/pkg/middleware"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	ginPrometheus "github.com/zsais/go-gin-prometheus"
)

var _ Handler = (*restHandler)(nil)

type Handler interface {
	RegisterHandler(engine *gin.Engine) error
}

func NewRestHandler(uc service.UseCase, swaggerAddr string, prom *ginPrometheus.Prometheus) Handler {
	return &restHandler{
		useCase:     uc,
		swaggerHost: swaggerAddr,
		ginProm:     prom,
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
	ginProm     *ginPrometheus.Prometheus
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

func (c *restHandler) prometheus(engine *gin.Engine) {
	paramSet := map[string]string{
		"name": ":name",
	}

	c.ginProm.ReqCntURLLabelMappingFn = func(c *gin.Context) string {
		url := c.Request.URL.Path
		for _, param := range c.Params {
			if value, exist := paramSet[param.Key]; exist {
				url = strings.Replace(url, param.Value, value, 1)
				break
			}
		}
		return url
	}
	c.ginProm.Use(engine)
}

func (c *restHandler) jwt(secret string) (*jwt.GinJWTMiddleware, error) {
	const identityKey = "userName"
	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "game",
		Key:         []byte(secret),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour * 24,
		IdentityKey: identityKey,
		LoginResponse: func(ctx *gin.Context, code int, token string, expire time.Time) {
			c.ResponseWithData(ctx, &internal.LoginRsp{
				Token:    token,
				ExpireIn: int(expire.Sub(time.Now()) / time.Second),
			})
		},
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*entity.User); ok {
				return jwt.MapClaims{
					identityKey: v.UserName,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &entity.User{
				UserName: claims[identityKey].(string),
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			req := &internal.LoginReq{}
			if err := c.ShouldBind(req); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			userID := req.UserName
			password := req.Password

			if (userID == "admin" && password == "admin") || (userID == "test" && password == "test") {
				return &entity.User{
					UserName: userID,
				}, nil
			}

			return nil, jwt.ErrFailedAuthentication
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			if v, ok := data.(*entity.User); ok && v.UserName == "admin" {
				return true
			}

			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		TokenLookup:   "header: Authorization, query: token, cookie: jwt",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	})
}

func (c *restHandler) RegisterHandler(engine *gin.Engine) error {
	c.healthCheck(engine)
	c.swaggerDocs(engine)
	c.prometheus(engine)
	jwtMiddle, err := c.jwt("test")
	if err != nil {
		return err
	}

	group1 := engine.Group("/svr/v1")
	group1.Use(middleware.Logger())
	group1.GET("hello/:name", c.Hello)
	group1.POST("hello/:name", c.Hello)
	group1.POST("login", jwtMiddle.LoginHandler)
	group1.GET("/refresh-token", jwtMiddle.RefreshHandler)
	group1.Use(jwtMiddle.MiddlewareFunc())
	{
		group1.GET("/biz", c.Hello)
	}

	return nil
}
