package server

import (
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"go-template/internal/base/conf"
	httpHandlers "go-template/internal/handler/http"
	"go-template/internal/handler/http/response"
	"go-template/internal/router"
	"go-template/pkg/consts"
	"net/http"
)

type HttpServer struct {
	config        *conf.Config
	Echo          *echo.Echo
	apiRouter     *router.Api
	swaggerRouter *router.SwaggerRouter
	middleware    *httpHandlers.Middleware
}

// NewHTTPServer new http server.
func NewHTTPServer(
	config *conf.Config,
	apiRouter *router.Api,
	swaggerRouter *router.SwaggerRouter,
	middleware *httpHandlers.Middleware,
) *HttpServer {

	e := echo.New()
	e.Use(echoMiddleware.Recover())
	e.Use(middleware.Logger.Handler())
	e.Use(middleware.JSONResponse.Handler())

	return &HttpServer{
		config:        config,
		Echo:          e,
		apiRouter:     apiRouter,
		swaggerRouter: swaggerRouter,
		middleware:    middleware,
	}
}

func (hs *HttpServer) AllowAllCors() {
	var defaultCORSConfig = echoMiddleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "X-Requested-With", "X-Auth-Token", "Authorization"},
		AllowCredentials: true,
		AllowMethods:     []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		MaxAge:           12 * 60,
	}

	hs.Echo.Use(echoMiddleware.CORSWithConfig(defaultCORSConfig))
}

func (hs *HttpServer) BizRouter() *echo.Echo {
	hs.AllowAllCors()

	rootGroup := hs.Echo.Group("")

	// swagger
	hs.swaggerRouter.Register(rootGroup)

	apiV1 := rootGroup.Group("/api/v1")
	{
		//apiV1.Use(corsMiddleWare)
		apiV1.Use(hs.middleware.JSONResponse.Handler())
		apiV1.Use(hs.middleware.Auth.Handler())
		hs.apiRouter.InitApiRouter(apiV1)
	}

	apiV1NoAuth := rootGroup.Group("/api/v1")
	{
		//apiV1.Use(corsMiddleWare)
		hs.apiRouter.InitNoAuthApiRouter(apiV1NoAuth)
	}

	hs.Echo.RouteNotFound("/*", func(ctx echo.Context) error {
		return response.Ctx(ctx).Status(http.StatusNotFound).Error(consts.ErrPageNotFound).Send()
	})

	return hs.Echo
}

func (hs *HttpServer) MetricRouter() *echo.Echo {
	e := echo.New()
	e.Use(echoMiddleware.Recover())

	metricGroup := e.Group("")
	// prometheus
	metricGroup.Use(echoprometheus.NewMiddleware(hs.config.App.Name))

	metricGroup.GET("/metrics", echoprometheus.NewHandler())

	metricGroup.GET("/healthz", func(ctx echo.Context) error { return ctx.String(200, "OK") })

	return e
}
