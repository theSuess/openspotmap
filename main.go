package main

import (
	"crypto/tls"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	apiV0 "github.com/theSuess/openspotmap/api/v0"
)

func main() {
	port := os.Getenv("PORT")
	dburl := os.Getenv("DATABASE_URL")

	e := echo.New()
	e.Use(middleware.Logger())

	e.Logger.SetLevel(log.DEBUG)

	db, err := pgx.Connect(connStringToConfig(dburl))
	if err != nil {
		e.Logger.Fatal(err)
		return
	}
	defer db.Close()

	v0 := apiV0.New(db)

	api := e.Group("/api", addCORSHeader)

	v0router := api.Group("/v0")
	v0router.GET("/spots", v0.GetSpots)
	v0router.POST("/spots", v0.AddSpot, v0.Authenticate("create"), v0.SpotFromBody())
	v0router.GET("/spots/:id", v0.GetSpot, v0.InjectSpot())
	v0router.DELETE("/spots/:id", v0.DeleteSpot, v0.Authenticate("delete"), v0.InjectSpot())
	v0router.PUT("/spots/:id", v0.UpdateSpot, v0.Authenticate("update"), v0.InjectSpot(), v0.SpotFromBody())
	v0router.PUT("/spots/:id", v0.UpdateSpot, v0.Authenticate("update"), v0.InjectSpot(), v0.SpotFromBody())
	v0router.OPTIONS("/spots",func(c echo.Context) error { return c.NoContent(200) })
    v0router.OPTIONS("/spots/:id",func(c echo.Context) error { return c.NoContent(200) })

	e.Start(":" + port)
}

func addCORSHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Add("Access-Control-Allow-Origin", "*")
		c.Response().Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept,X-API-Key")
		c.Response().Header().Add("Access-Control-Allow-Methods", "PUT,GET")
		return next(c)
	}
}

func connStringToConfig(c string) pgx.ConnConfig {
	c = c[11:] // remove "postgres://"
	splt := strings.Split(c, "@")
	spltCred := strings.Split(splt[0], ":")
	user := spltCred[0]
	password := spltCred[1]

	spltDB := strings.Split(splt[1], "/")
	db := spltDB[1]
	host, ports, _ := net.SplitHostPort(spltDB[0])
	port, _ := strconv.Atoi(ports)
	return pgx.ConnConfig{
		Host:      host,
		Port:      uint16(port),
		Database:  db,
		User:      user,
		Password:  password,
		TLSConfig: &tls.Config{InsecureSkipVerify: true}, //TLSConfig must be set to use SSL
	}
}
