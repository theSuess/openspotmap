package main

import (
	"database/sql"
	"os"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	apiV0 "github.com/theSuess/openspotmap/api/v0"
)

type Point struct {
	Longitude float64
	Latitude  float64
}
type Spot struct {
	Name        string
	Description string
	Location    Point
	Images      []string
}

func main() {
	port := os.Getenv("PORT")
	dburl := os.Getenv("DATABASE_URL")

	e := echo.New()
	e.Use(middleware.Logger())

	db, err := sql.Open("pgx", dburl)
	if err != nil {
		e.Logger.Fatal(err)
		return
	}
	defer db.Close()

	v0 := apiV0.New(db)

	api := e.Group("/api")
	v0router := api.Group("/v0")
	v0router.GET("/spots", v0.GetSpots)

	e.Start(":" + port)
}
