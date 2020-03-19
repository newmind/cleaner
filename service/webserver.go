package service

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

var (
	e *echo.Echo
)

func StartWebServer(port string) {
	log.Infoln("Starting HTTP service at " + port)

	e = echo.New()

	// Enable metrics middleware
	// "github.com/labstack/echo-contrib/prometheus"
	//p := prometheus.NewPrometheus("echo", nil)
	//p.Use(e)

	//metrics.CctvConnected.Set(0)
	//e.Use(echoprometheus.MetricsMiddleware())
	//e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	//e.GET("/health", echo.WrapHandler(promhttp.Handler()))

	e.GET("/", hello)
	e.GET("/health", healthCheck)

	e.Logger.Fatal(e.Start(":" + port))
}

func StopWebServer() {
	if e != nil {
		e.Shutdown(context.Background())
	}
}

func healthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
