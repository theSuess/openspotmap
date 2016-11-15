package api

import (
	"github.com/labstack/echo"
	"net/http"
)

func (api *api) GetSpots(c echo.Context) error {
	rows, err := api.db.Query("SELECT name,description,ST_X(location::geometry) as longitude,ST_Y(location::geometry) as latitude,images FROM spots")
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
}
