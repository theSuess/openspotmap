package api

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
