package vector

import (
	"geo"
	"math"
)


//y = lat
//x = lng
func Degree(p1, p2 geo.Point) float64{

	xdiff := p2.Lng - p1.Lng
	ydiff := p2.Lat - p1.Lat

	rad := math.Atan(ydiff / xdiff)
	deg := rad * 180.0 / math.Pi
	switch {
	case xdiff >= 0 && ydiff >= 0:
		break
	case xdiff <= 0 && ydiff >= 0:
		deg = 180 + deg
		break
	case xdiff <=0 && ydiff <= 0:
		deg = 180 + deg
		break
	case xdiff >= 0 && ydiff <= 0:
		deg = 360 + deg
		break
	}

	return deg
}