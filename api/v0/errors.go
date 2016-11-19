package api

import (
	"fmt"
)

var (
	ErrInternal     = APIError{Response: Response{Code: 500}, Message: "Internal Server Error"}
	ErrSpotNotFound = APIError{Response: Response{Code: 404}, Message: "Spot not found."}
)

func errorMustBeType(v string, t string) APIError {
	return APIError{Response: Response{Code: 400}, Message: fmt.Sprintf("%s must be of type %s", v, t)}
}

func errorMustBe(v string, t string) APIError {
	return APIError{Response: Response{Code: 400}, Message: fmt.Sprintf("%s must be %s", v, t)}
}
