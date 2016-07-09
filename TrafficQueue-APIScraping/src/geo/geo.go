package geo

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