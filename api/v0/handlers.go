package api

import (
	"github.com/jackc/pgx"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

func (api *api) GetSpots(c echo.Context) error {
	rows, err := api.db.Query("SELECT id,name,description,ST_X(location::geometry) as longitude,ST_Y(location::geometry) as latitude,images FROM spots")
	if err != nil {
		return err
	}
	var res []Spot
	for rows.Next() {
		var spot Spot
		rows.Scan(&spot.Id, &spot.Name, &spot.Description, &spot.Location.Longitude, &spot.Location.Latitude, &spot.Images)
		res = append(res, spot)
	}
	return c.JSON(http.StatusOK, SpotList{Response: Response{Type: "result", Code: 200}, Spots: res, Length: len(res)})
}

func (api *api) GetSpot(c echo.Context) error {
	q := c.Param("id")
	if q == "" {
		return c.JSON(http.StatusBadRequest, APIError{Response: Response{Type: "error", Code: 400}, Message: "Spot ID must be a specified"})
	}
	id, err := strconv.Atoi(q)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Response: Response{Type: "error", Code: 400}, Message: "Spot ID must be a number"})
	}
	var spot Spot
	err = api.db.QueryRow(`SELECT id,name,description,ST_X(location::geometry) as longitude,ST_Y(location::geometry) as latitude,images
                            FROM spots
                            WHERE id = $1`, id).Scan(&spot.Id, &spot.Name, &spot.Description, &spot.Location.Longitude, &spot.Location.Latitude, &spot.Images)
	switch err {
	case pgx.ErrNoRows:
		return c.JSON(http.StatusNotFound, ErrSpotNotFound)
	case nil:
		return c.JSON(http.StatusOK, SpotResponse{Response: Response{Type: "result", Code: 200}, Spot: spot})
	default:
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, ErrInternal)
	}
}
