package main

import (

)
import (
	"geo"
	"fmt"
	"google"
	"net/http"
	"log"
	"io/ioutil"
	"encoding/json"
	"math/rand"
	"gopkg.in/mgo.v2"
	"mongo"
	"gopkg.in/mgo.v2/bson"
	"time"
	"lg"
	"model"
	"runtime"
)

const phuket_start_lat = 7.755442
const phuket_stop_lat = 8.19947
const phuket_start_lng = 98.256715
const phuket_stop_lng = 98.444073

const start_lat = phuket_start_lat
const stop_lat = phuket_stop_lat
const start_lng = phuket_start_lng
const stop_lng = phuket_stop_lng

const column_count = 5

func main(){
	runtime.GOMAXPROCS(runtime.NumCPU())

	var phuketGeoLimit = geo.GeoLimitSquare{
		StartLat:start_lat,
		StopLat:stop_lat,
		StartLng:start_lng,
		StopLng:stop_lng,
	}
	phuketGeoLimit.FindDiffLatLng()
	var currentGeoRegion = phuketGeoLimit

	ra := rand.New(rand.NewSource(time.Now().UnixNano()))

	session, err := mgo.Dial("127.0.0.1")
	session.SetMode(mgo.Monotonic, true)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	mongo.AddLimitedRegion(session, currentGeoRegion)
	nodes := createNodeRegion(currentGeoRegion, session)
	channel := make(chan int, len(nodes))
	for i, node := range nodes{
		go webScraping2(i, node, ra, channel)
	}

	for ;;{
		select {
		case r := <- channel:
			fmt.Println("response from node ", r)
		}
	}
}

func createNodeRegion(currentGeoRegion geo.GeoLimitSquare, session *mgo.Session) []geo.GeoLimitSquare{
	nodeWidth := currentGeoRegion.LngDiff / float64(column_count)
	cc := session.DB("GoTrafficQueue").C("region");
	cc.RemoveAll(bson.M{"big":0})
	var nodes []geo.GeoLimitSquare
	for i := 0 ; i < column_count ; i++{
		c := currentGeoRegion
		for j, k := 0, c.StartLat ; k <= c.StopLat ; k+=nodeWidth{
			node := geo.GeoLimitSquare{
				StartLng:c.StartLng + (float64(i)*nodeWidth),
				StopLng:c.StartLng + (float64(i+1)*nodeWidth),
				StartLat:c.StartLat + (float64(j)*nodeWidth),
				StopLat:c.StartLat + (float64(j+1)*nodeWidth),
			}
			nodes = append(nodes, node)
			cc.Insert(bson.M{"big":0, "geo":node})
			j++
		}
	}
	return nodes
}

func regisNodeWorker(nodeNum int, currentGeoRegion geo.GeoLimitSquare, session *mgo.Session){
	c := session.DB("GoTrafficQueue").C("worker")
	c.Upsert(bson.M{"node":nodeNum}, bson.M{"node":nodeNum, "limit":currentGeoRegion})
}

func webScraping2 (nodeNum int, currentGeoRegion geo.GeoLimitSquare, ra *rand.Rand, ch chan int){
	session, _ := mgo.Dial("127.0.0.1")
	regisNodeWorker(nodeNum, currentGeoRegion, session)
	c := session.DB("GoTrafficQueue").C("road_relate")
	lc := session.DB("GoTrafficQueue").C("visualize")
	pc := session.DB("GoTrafficQueue").C("polyline")
	for ;true; {
		rand1 := randomPosition(currentGeoRegion, *ra)
		rand2 := randomPosition(currentGeoRegion, *ra)
		grr := requestGoogleRouteApi(rand1, rand2)
		lg.PrintLog("Position", fmt.Sprintf("%.5f->%.5f, %.5f->%.5f",rand2.Lat, rand1.Lat, rand2.Lng, rand1.Lng))
		if grr.Status != "OK"{
			lg.PrintLog("ERR", "STATUS=" + grr.Status)
			continue
		}

		lg.PrintLog("Route", fmt.Sprintf("Found %d Routes", len(grr.Routes)))
		for i := 0 ; i < len(grr.Routes); i++{
			pc.Insert(bson.M{"polyline":grr.Routes[i].OverviewPolyline.Points})
			lc.Upsert(bson.M{"node":nodeNum}, bson.M{"node":nodeNum, "polyline":grr.Routes[i].OverviewPolyline.Points})

			roadLatLng := google.DecodePolyline(grr.Routes[i].OverviewPolyline.Points)

			for j:=1;j<len(roadLatLng)-2;j++{
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
			}
		}
		ch <- nodeNum
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

	if lp.LatDiff == 0 || lp.LngDiff == 0{
		lp.FindDiffLatLng()
	}

	r.Lat = ra.Float64()
	r.Lng = ra.Float64()

	r.Lat = lp.StartLat + (r.Lat * lp.LatDiff)
	r.Lng = lp.StartLng + (r.Lng * lp.LngDiff)
	return r
}