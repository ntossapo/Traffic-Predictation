package main

import (

)
import (
	"geo"
	"time"
	"lg"
	"fmt"
	"google"
	"net/http"
	"log"
	"io/ioutil"
	"encoding/json"
	"math/rand"
	"math"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"model"
)

const phuket_start_lat = 7.755442
const phuket_stop_lat = 8.19947
const phuket_start_lng = 98.256715
const phuket_stop_lng = 98.444073

const start_lat = phuket_start_lat
const stop_lat = phuket_stop_lat
const start_lng = phuket_start_lng
const stop_lng = phuket_stop_lng

func main(){
	var phuketGeoLimit = geo.GeoLimitSquare{
		StartLat:start_lat,
		StopLat:stop_lat,
		StartLng:start_lng,
		StopLng:stop_lng,
	}

	ra := rand.New(rand.NewSource(time.Now().UnixNano()))

	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		log.Fatal(err)
	}
	session.SetMode(mgo.Monotonic, true)
	c := session.DB("GoTrafficQueue").C("road_relate")
	lc := session.DB("GoTrafficQueue").C("visualize")
	pc := session.DB("GoTrafficQueue").C("polyline")
	for ;true; {

		rand1 := randomPosition(phuketGeoLimit, *ra)
		rand2 := randomPosition(phuketGeoLimit, *ra)
		grr := requestGoogleRouteApi(rand1, rand2)
		lg.PrintLog("Position", fmt.Sprintf("%.5f->%.5f, %.5f->%.5f",rand2.Lat, rand1.Lat, rand2.Lng, rand1.Lng))
		if grr.Status != "OK"{
			lg.PrintLog("ERR", "STATUS=" + grr.Status)
			continue
		}

		lg.PrintLog("Route", fmt.Sprintf("Found %d Routes", len(grr.Routes)))
		for i := 0 ; i < len(grr.Routes); i++{
			pc.Insert(bson.M{"polyline":grr.Routes[i].OverviewPolyline.Points})
			lc.RemoveAll(bson.M{})
			lc.Insert(bson.M{"last":grr.Routes[i].OverviewPolyline.Points})

			roadLatLng := google.DecodePolyline(grr.Routes[i].OverviewPolyline.Points)
			for j:=0;j<len(roadLatLng)-1;j++{
				foundModel := &model.Model{}
				foundModel.Host = roadLatLng[j]
				err := c.Find(bson.M{
					"host":bson.M{
						"lat":roadLatLng[j].Lat,
						"lng":roadLatLng[j].Lng,
					}}).One(&foundModel)
				if err != nil {
					if err.Error() == "not found"{
						lg.PrintLog("New Record", fmt.Sprintf("Add new Data %.5f, %.5f", roadLatLng[j].Lat, roadLatLng[j].Lng))
						newModel := &model.Model{}
						newModel.NewInstance(roadLatLng[j], nil)
						newModel.Append(roadLatLng[j+1])
						err := c.Insert(newModel)
						if err != nil {
							lg.PrintLog("Error", err.Error())
						}
					}
				}else{
					lg.PrintLog("Found Record", fmt.Sprintf("Found Data %.5f, %.5f", roadLatLng[j].Lat, roadLatLng[j].Lng))
					if !foundModel.ContainParent(roadLatLng[j+1]){
						lg.PrintLog("Found Record", fmt.Sprintf("Add relation %.5f, %.5f to %.5f, %.5f",
							roadLatLng[j+1].Lat, roadLatLng[j+1].Lng,
							roadLatLng[j].Lat, roadLatLng[j].Lng,
						))
						oldModel := foundModel.CopyInstance()
						oldModel.Append(roadLatLng[j+1])
						err := c.Update(foundModel, oldModel)
						if err != nil {
							lg.PrintLog("Error", err.Error())
						}
					}else{
						lg.PrintLog("Already Relation", "Found Already Relation Road")
					}
				}

				//lg.PrintLog("Point", fmt.Sprintf("%.5f, %.5f", roadLatLng[j].Lat, roadLatLng[j].Lng))

			}
		}


		time.Sleep(time.Second * 10)
	}
}

func requestGoogleRouteApi(r1 geo.RandomPoint, r2 geo.RandomPoint) google.GoogleRouteRequest{
	url := fmt.Sprintf("https://maps.googleapis.com/maps/api/directions/json?origin=%.5f,%.5f&destination=%.5f,%.5f",
		r1.Lat, r1.Lng, r2.Lat, r2.Lng)

	req, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	byteData, err := ioutil.ReadAll(req.Body)
	if err != nil{
		log.Fatal(err)
	}
	defer req.Body.Close()

	result := google.GoogleRouteRequest{}

	json.Unmarshal(byteData, &result)

	return result
}

func randomPosition(lp geo.GeoLimitSquare, ra rand.Rand) geo.RandomPoint{
	var r geo.RandomPoint;


	lp = findDiffLatLng(lp)
	r.Lat = ra.Float64()
	r.Lng = ra.Float64()

	r.Lat = lp.StartLat + (r.Lat * lp.LatDiff)
	r.Lng = lp.StartLng + (r.Lng * lp.LngDiff)
	return r
}

func findDiffLatLng(g geo.GeoLimitSquare) geo.GeoLimitSquare{
	g.LatDiff = math.Abs(g.StartLat - g.StopLat)
	g.LngDiff = math.Abs(g.StartLng - g.StopLng)
	return  g
}