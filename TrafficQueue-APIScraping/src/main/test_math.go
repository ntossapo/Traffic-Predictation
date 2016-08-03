package main

import (
	"geo"
	"math"
	"fmt"
)

func main(){
	fmt.Println(degree(geo.Point{Lat:8.2269, Lng:98.35612}, geo.Point{Lat:8.22732, Lng:98.35534}))
	fmt.Println(degree(geo.Point{Lat:8.2269, Lng:98.35612}, geo.Point{Lat:8.22587, Lng:98.35804}))
}


//lat = y
//lng = x
func degree(p1, p2 geo.Point) float64{
	var q int
	x := p2.Lng - p1.Lng
	y := p2.Lat - p1.Lat

	switch {
	case x >= 0 && y >= 0:
		q = 1
		break
	case x <= 0 && y >= 0:
		q = 2
		break
	case x <= 0 && y <= 0:
		q = 3
		break
	case x >= 0 && y <= 0:
		q = 4
		break
	}

	fmt.Println(q)

	deg := math.Atan((p2.Lat - p1.Lat) / (p2.Lng - p1.Lng)) * 180.0 / math.Pi
	switch q {
	case 2:
		deg = 180 + deg
		break
	case 3:
		deg = 180 + deg
		break
	case 4:
		deg = 360 + deg
		break
	}
	return deg
}
