package api

import (
	"github.com/labstack/echo"
	"net/http"
	"regexp"
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
			return c.JSON(http.StatusBadRequest, errorMustBeType("limit", "integer"))
		}
		if limit <= 0 {
			return c.JSON(http.StatusBadRequest, errorMustBe("limit", "positive"))
		}
	}
	var offset int
	switch offsetString {
	case "":
		offset = 0
	default:
		offset, err = strconv.Atoi(offsetString)
		if err != nil {
			return c.JSON(http.StatusBadRequest, errorMustBeType("offset", "integer"))
		}
		if offset < 0 {
			return c.JSON(http.StatusBadRequest, errorMustBe("offset", "positive or zero"))
		}
	}

	var result SpotList
	nearString := c.QueryParam("near")

	if nearString != "" {
		nearString = strings.TrimSpace(nearString)
		correctFormat, err := regexp.MatchString(`^[0-9]+\.?[0-9]*,[0-9]+\.?[0-9]*$`, nearString)

		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, ErrInternal)
		}

		if !correctFormat {
			return c.JSON(http.StatusBadRequest, errorMustBeType("near", "float,float (latitude,longitude)"))
		}

		distanceString := c.QueryParam("distance")
		if distanceString == "" {
			distanceString = "5000"
		}
		distance, err := strconv.Atoi(distanceString)
		if err != nil {
			return c.JSON(http.StatusBadRequest, errorMustBeType("distance", "integer (in meters)"))
		}
		splt := strings.Split(nearString, ",")
		result, err = api.getSpotsNear(limit, offset, splt[1], splt[0], distance)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, ErrInternal)
		}
	} else {
		result, err = api.getActiveSpots(limit, offset)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, ErrInternal)
		}
	}
	return c.JSON(http.StatusOK, result)
}

func (api *api) GetSpot(c echo.Context) error {
	spot := c.Get("spot").(Spot)
	return c.JSON(http.StatusOK, SpotResponse{Response: Response{Code: 200}, Spot: spot})
}

func (api *api) AddSpot(c echo.Context) error {
	key := c.Get("key").(string)
	s := c.Get("reqspot").(Spot)

	err := api.db.QueryRow(`SELECT 1 FROM spots WHERE name=$1`, s.Name).Scan(nil)
	if err == nil {
		return c.JSON(http.StatusBadRequest, errorGeneral(http.StatusBadRequest, "A spot with that name already exists"))
	}

	err = api.db.QueryRow(`SELECT 1 FROM spots WHERE location=ST_MakePoint($1,$2)`, s.Location.Longitude, s.Location.Latitude).Scan(nil)
	if err == nil {
		return c.JSON(http.StatusBadRequest, errorGeneral(http.StatusBadRequest, "A spot with that location already exists"))
	}

	trans, err := api.db.Begin() // Starting transaction
	_, err = trans.Exec(`INSERT INTO spots (name,description,location,images,submitter) VALUES
                        ($1,$2,ST_MakePoint($3,$4),$5,$6)`,
		s.Name, s.Description, s.Location.Longitude, s.Location.Latitude, s.Images, key)

	if err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, ErrInternal)
	}

	_, err = trans.Exec(`INSERT INTO activespots VALUES ((SELECT id FROM spots WHERE name=$1))`, s.Name)
	if err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, ErrInternal)
	}
	err = trans.Commit()
	if err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, ErrInternal)
	}
	return c.NoContent(http.StatusCreated)
}

func (api *api) DeleteSpot(c echo.Context) error {
	spot := c.Get("spot").(Spot)
	api.db.Exec(`DELETE FROM activespots WHERE id=$1`, spot.Id)
	return c.NoContent(http.StatusNoContent)
}

func (api *api) UpdateSpot(c echo.Context) error {
	spot := c.Get("spot").(Spot)
	s := c.Get("reqspot").(Spot)
	tr, err := api.db.Begin()

	if err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, ErrInternal)
	}
	if s.Name != "" {
		tr.Exec(`UPDATE spots SET name = $1 WHERE id=$2`, s.Name, spot.Id)
	}
	if s.Description != "" {
		tr.Exec(`UPDATE spots SET description = $1 WHERE id=$2`, s.Description, spot.Id)
	}
	if s.Images != nil {
		if c.QueryParam("imgreplace") != "" {
			tr.Exec(`UPDATE spots SET images = $1 WHERE id=$2`, s.Images, spot.Id)
		} else {
			tr.Exec(`UPDATE spots SET images = images || $1 WHERE id=$2`, s.Images, spot.Id)
		}
	}
	if s.Location.Latitude != 0 {
		tr.Exec(`UPDATE spots
               SET location = ST_MakePoint((SELECT ST_X(location::geometry) FROM spots WHERE id = $2),$1)
               WHERE id=$2`,
			s.Location.Latitude, spot.Id)
	}
	if s.Location.Longitude != 0 {
		tr.Exec(`UPDATE spots
               SET location = ST_MakePoint($1,(SELECT ST_Y(location::geometry) FROM spots WHERE id = $2))
               WHERE id=$2`,
			s.Location.Longitude, spot.Id)
	}

	err = tr.Commit()
	if err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, ErrInternal)
	}
	return c.NoContent(http.StatusOK)
}

func (api *api) getActiveSpots(limit int, offset int) (SpotList, error) {
	var totalCount int
	err := api.db.QueryRow("SELECT count(*) FROM activespots").Scan(&totalCount)
	if err != nil {
		return SpotList{}, err
	}

	rows, err := api.db.Query(`SELECT id,name,description,ST_X(location::geometry),ST_Y(location::geometry),images FROM activespots NATURAL JOIN spots ORDER BY id LIMIT $1 OFFSET $2`, limit, offset)
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
		Response: Response{Code: 200},
		Spots:    res, Length: len(res),
		Next:  offset + len(res),
		Total: totalCount,
	}, nil
}

func (api *api) getSpotsNear(limit int, offset int, long string, lat string, distance int) (SpotList, error) {
	distFloat := float64(distance)
	var totalCount int
	err := api.db.QueryRow("SELECT count(*) FROM activespots NATURAL JOIN spots WHERE ST_DISTANCE(ST_MakePoint($1,$2),location) <= $3", long, lat, distFloat).Scan(&totalCount)
	if err != nil {
		return SpotList{}, err
	}

	rows, err := api.db.Query(
		"SELECT id,name,description,ST_X(location::geometry),ST_Y(location::geometry),images FROM spots NATURAL JOIN spots WHERE ST_DISTANCE(ST_MakePoint($1,$2),location) < $3 ORDER BY id LIMIT $4 OFFSET $5",
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
		Response: Response{Code: 200},
		Spots:    res, Length: len(res),
		Next:  offset + len(res),
		Total: totalCount,
	}, nil
}
