package api

var (
	ErrInternal     = APIError{Response: Response{Type: "error", Code: 500}, Message: "Internal Server Error"}
	ErrSpotNotFound = APIError{Response: Response{Type: "error", Code: 404}, Message: "Spot not found."}
)
