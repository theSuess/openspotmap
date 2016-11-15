package main

import (
	"database/sql"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
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

	e.GET("/", func(c echo.Context) error {
		rows, err := db.Query("SELECT name,description,ST_X(location::geometry) as longitude,ST_Y(location::geometry) as latitude,images FROM spots")
		if err != nil {
			return err
		}
		var res []Spot
		for rows.Next() {
			var spot Spot
			rows.Scan(&spot.Name, &spot.Description, &spot.Location.Longitude, &spot.Location.Latitude, &spot.Images)
			res = append(res, spot)
		}
		return c.JSON(http.StatusOK, res)
	})
	e.Start(":" + port)
}
