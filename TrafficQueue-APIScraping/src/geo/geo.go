package geo

import "math"

type GeoLimitSquare struct{
	StartLat	float64
	StopLat		float64
	StartLng	float64
	StopLng		float64

	LatDiff		float64
	LngDiff		float64
}

type Point struct {
	Lat		float64
	Lng		float64
}

type RandomPoint Point

func (g *GeoLimitSquare) FindDiffLatLng(){
	g.LatDiff = math.Abs(g.StartLat - g.StopLat)
	g.LngDiff = math.Abs(g.StartLng - g.StopLng)
}