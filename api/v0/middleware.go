package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jackc/pgx"
	"github.com/labstack/echo"
)

func (api *api) Authenticate(level string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := c.Request().Header.Get("X-API-Key")
			if key == "" {
				return c.JSON(http.StatusBadRequest, errorMustBe("X-API-Key header", "specified"))
			}

			var permissions []string
			err := api.db.QueryRow(`SELECT permissions FROM keys WHERE id = $1`, key).Scan(&permissions)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, errorGeneral(http.StatusUnauthorized, "Invalid Key"))
			}
			if !stringInSlice(level, permissions) {
				return c.JSON(http.StatusForbidden, errorGeneral(http.StatusForbidden, "Your misses the following permissions: '"+level+"'"))
			}
			c.Set("key", key)
			return next(c)
		}
	}
}

func (api *api) InjectSpot() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			q := c.Param("id")
			if q == "" {
				return c.JSON(http.StatusBadRequest, errorMustBe("Spot ID", "specified"))
			}
			id, err := strconv.Atoi(q)
			if err != nil {
				return c.JSON(http.StatusBadRequest, errorMustBeType("Spot ID", "integer"))
			}
			var spot Spot
			err = api.db.QueryRow(`SELECT id,name,description,ST_X(location::geometry) as longitude,ST_Y(location::geometry) as latitude,images
                            FROM spots
                            WHERE id = $1`, id).Scan(&spot.Id, &spot.Name, &spot.Description, &spot.Location.Longitude, &spot.Location.Latitude, &spot.Images)
			switch err {
			case pgx.ErrNoRows:
				return c.JSON(http.StatusNotFound, ErrSpotNotFound)
			case nil:
				c.Set("spot", spot)
				return next(c)
			default:
				c.Logger().Error(err)
				return c.JSON(http.StatusInternalServerError, ErrInternal)
			}
		}
	}
}

func (api *api) SpotFromBody() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			decoder := json.NewDecoder(req.Body)
			var spot Spot
			err := decoder.Decode(&spot)
			if err != nil {
				return c.JSON(http.StatusBadRequest, errorMustBe("Request Body", "valid spot json"))
			}
			defer req.Body.Close()
			c.Set("reqspot", spot)
			return next(c)
		}
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
