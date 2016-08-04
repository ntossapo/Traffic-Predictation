package main

import (
	//"math"
	"fmt"
	"math"
)

func main(){
	fmt.Println(isSameVector(0, 15) == true)
	fmt.Println(isSameVector(0, 16) == false)
	fmt.Println(isSameVector(0, 17) == false)
	fmt.Println(isSameVector(0, 355) == true)
	fmt.Println(isSameVector(355, 0) == true)
	fmt.Println(isSameVector(270, 283) == true)
	fmt.Println(isSameVector(8, 350) == false)
}

//func isSameVector(deg1, deg2 float64) bool{
//	if math.Abs(deg1-deg2) <= 15{
//		return true
//	}else if math.Abs((360+deg1) - deg2) <= 15{
//		return true
//	}else if math.Abs((360+deg2) - deg1) <= 15{
//		return true
//	}
//	return false
//}