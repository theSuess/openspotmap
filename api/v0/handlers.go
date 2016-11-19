package api

import (
	"github.com/jackc/pgx"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
	"strings"
)

func (api *api) GetSpots(c echo.Context) error {
	limitString := c.QueryParam("limit")
	offsetString := c.QueryParam("offset")
	var err error
	var limit int
	switch limitString {
	case "":
		limit = 100
	default:
		limit, err = strconv.Atoi(limitString)
		if err != nil {
			return c.JSON(http.StatusBadRequest, APIError{Response: Response{Type: "error", Code: 400}, Message: "Limit must be a number"})
		}
		if limit <= 0 {
			return c.JSON(http.StatusBadRequest, APIError{Response: Response{Type: "error", Code: 400}, Message: "Limit must be positive"})
		}
	}
	var offset int
	switch offsetString {
	case "":
		offset = 0
	default:
		offset, err = strconv.Atoi(offsetString)
		if err != nil {
			return c.JSON(http.StatusBadRequest, APIError{Response: Response{Type: "error", Code: 400}, Message: "Offset must be a number"})
		}
		if offset < 0 {
			return c.JSON(http.StatusBadRequest, APIError{Response: Response{Type: "error", Code: 400}, Message: "Offset must be positive or 0"})
		}
	}

	var result SpotList
	nearString := c.QueryParam("near")

	if nearString != "" {
		distanceString := c.QueryParam("distance")
		distance, err := strconv.Atoi(distanceString)
		if err != nil {
			return c.JSON(http.StatusBadRequest, APIError{Response: Response{Type: "error", Code: 400}, Message: "Distance must be a number"})
		}
		splt := strings.Split(nearString, ",")
		result, err = api.getSpotsNear(limit, offset, splt[1], splt[0], distance)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, ErrInternal)
		}
	} else {
		result, err = api.getAllSpots(limit, offset)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, ErrInternal)
		}
	}
	return c.JSON(http.StatusOK, result)
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

func (api *api) getAllSpots(limit int, offset int) (SpotList, error) {
	var totalCount int
	err := api.db.QueryRow("SELECT count(*) FROM spots").Scan(&totalCount)
	if err != nil {
		return SpotList{}, err
	}

	rows, err := api.db.Query("SELECT id,name,description,ST_X(location::geometry),ST_Y(location::geometry),images FROM spots ORDER BY id LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return SpotList{}, err
	}
	var res []Spot
	for rows.Next() {
		var spot Spot
		rows.Scan(&spot.Id, &spot.Name, &spot.Description, &spot.Location.Longitude, &spot.Location.Latitude, &spot.Images)
		res = append(res, spot)
	}
	return SpotList{
		Response: Response{Type: "result", Code: 200},
		Spots:    res, Length: len(res),
		Next:  offset + len(res),
		Total: totalCount,
	}, nil
}

func (api *api) getSpotsNear(limit int, offset int, long string, lat string, distance int) (SpotList, error) {
	distFloat := float64(distance)
	var totalCount int
	err := api.db.QueryRow("SELECT count(*) FROM spots WHERE ST_DISTANCE(ST_MakePoint($1,$2),location) <= $3", long, lat, distFloat).Scan(&totalCount)
	if err != nil {
		return SpotList{}, err
	}

	rows, err := api.db.Query(
		"SELECT id,name,description,ST_X(location::geometry),ST_Y(location::geometry),images FROM spots WHERE ST_DISTANCE(ST_MakePoint($1,$2),location) < $3 ORDER BY id LIMIT $4 OFFSET $5",
		long,
		lat,
		distFloat,
		limit,
		offset)
	if err != nil {
		return SpotList{}, err
	}
	var res []Spot
	for rows.Next() {
		var spot Spot
		rows.Scan(&spot.Id, &spot.Name, &spot.Description, &spot.Location.Longitude, &spot.Location.Latitude, &spot.Images)
		res = append(res, spot)
	}
	return SpotList{
		Response: Response{Type: "result", Code: 200},
		Spots:    res, Length: len(res),
		Next:  offset + len(res),
		Total: totalCount,
	}, nil
}