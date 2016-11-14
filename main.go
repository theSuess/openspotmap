package main

import (
	"net/http"
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
)

func main() {
	port := os.Getenv("PORT")
	e := echo.New()
	e.Use(middleware.Logger())
	e.SetLogLevel(0)

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello World")
	})
	e.Run(standard.New(":" + port))
}
