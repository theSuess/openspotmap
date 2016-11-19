package api

type Point struct {
	Longitude float64
	Latitude  float64
}
type Spot struct {
	Id          int
	Name        string
	Description string
	Location    Point
	Images      []string
}

type APIError struct {
	Response
	Message string
}

type Response struct {
	Code int
}

type SpotList struct {
	Response
	Length int
	Next   int
	Total  int
	Spots  []Spot
}
type SpotResponse struct {
	Response
	Spot Spot
}
