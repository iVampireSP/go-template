package server

import (
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	httpApi "go-template/internal/api/http"
	"go-template/internal/api/http/response"
	"go-template/internal/base/conf"
	"go-template/internal/base/logger"
	"go-template/internal/consts"
	"go-template/internal/router"
	"net/http"
	"strings"
)

type HttpServer struct {
	config        *conf.Config
	Fiber         *fiber.App
	apiRouter     *router.Api
	swaggerRouter *router.SwaggerRouter
	middleware    *httpApi.Middleware
}

// NewHTTPServer new http server.
func NewHTTPServer(
	config *conf.Config,
	apiRouter *router.Api,
	swaggerRouter *router.SwaggerRouter,
	middleware *httpApi.Middleware,
	logger *logger.Logger,
) *HttpServer {
	app := fiber.New(fiber.Config{
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			logger.Sugar.Errorf("fiber error: %s", err)
			return response.Ctx(ctx).Status(fiber.StatusInternalServerError).Error(consts.ErrInternalServerError).Send()
		},
	})
	app.Use(recover.New())
	app.Use(middleware.Logger.Handler())
	app.Use(middleware.JSONResponse.Handler())

	return &HttpServer{
		config:        config,
		Fiber:         app,
		apiRouter:     apiRouter,
		swaggerRouter: swaggerRouter,
		middleware:    middleware,
	}
}

func (hs *HttpServer) AllowAllCors() {
	if hs.config.Http.Cors.Enabled {
		var config = cors.Config{
			AllowOrigins:     strings.Join(hs.config.Http.Cors.AllowedOrigins, ","),
			AllowMethods:     strings.Join(hs.config.Http.Cors.AllowMethods, ","),
			AllowHeaders:     strings.Join(hs.config.Http.Cors.AllowHeaders, ","),
			AllowCredentials: hs.config.Http.Cors.AllowCredentials,
			ExposeHeaders:    strings.Join(hs.config.Http.Cors.ExposeHeaders, ","),
			MaxAge:           hs.config.Http.Cors.MaxAge,
		}

		hs.Fiber.Use(cors.New(config))
	}

}

func (hs *HttpServer) BizRouter() *fiber.App {
	hs.AllowAllCors()

	rootGroup := hs.Fiber.Group("")

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
		hs.apiRouter.InitNoAuthApiRouter(apiV1NoAuth)
	}

	// 404 Handler
	hs.Fiber.Use(func(ctx *fiber.Ctx) error {
		return response.Ctx(ctx).Status(fiber.StatusNotFound).Send()
	})

	return hs.Fiber
}

func (hs *HttpServer) MetricRouter() *fiber.App {
	app := fiber.New()

	app.Use(recover.New())

	metricGroup := app.Group("")

	prometheus := fiberprometheus.New(hs.config.App.Name)

	prometheus.RegisterAt(app, "/metrics")

	metricGroup.Use(prometheus.Middleware)

	metricGroup.Get("/healthz", func(ctx *fiber.Ctx) error {
		return ctx.Status(http.StatusOK).SendString("OK")
	})

	return app
}
