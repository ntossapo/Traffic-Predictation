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
	"model"
	"runtime"
	"utils/vector"
	"math"
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
	//for i := 0 ; i < 1 ; i ++{
		go webScraping2(i, node, ra, channel)
	}

	for ;;{
		select {
		case <- channel:
			//fmt.Println("response from node ", r)
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

//false do
//true not do
func  isSameVector(newPoint geo.Point, model model.Model) bool{
	result := false
	deg1 := vector.Degree(model.Host, newPoint)
	for i := 0 ; i < len(model.Parent) ; i++ {
		deg2 := vector.Degree(model.Host, model.Parent[i])
		fmt.Println("DEG", "deg1",deg1)
		fmt.Println("DEG", "deg2",deg2)
		fmt.Println("DEG", "Diff1",math.Abs(deg1-deg2))
		if math.Abs(deg1-deg2) <= 15{
			return true
		}else if math.Abs((360+deg1) - deg2) <= 15{
			fmt.Println("DEG", "Diff2", math.Abs((360+deg1) - deg2))
			return true
		}else if math.Abs((360+deg2) - deg1) <= 15{
			fmt.Println("DEG", "Diff3",math.Abs((360+deg2) - deg1))
			return true
		}
	}

	return result
}

func webScraping2 (nodeNum int, currentGeoRegion geo.GeoLimitSquare, ra *rand.Rand, ch chan int){
	session, _ := mgo.Dial("127.0.0.1")

	timeRand := 5 + rand.Intn(60 - 5)
	//fmt.Println("SLEEP", "Thread", nodeNum, "Sleep", timeRand, "Second")
	time.Sleep(time.Second * time.Duration(timeRand))

	regisNodeWorker(nodeNum, currentGeoRegion, session)
	c := session.DB("GoTrafficQueue").C("road_relate")
	lc := session.DB("GoTrafficQueue").C("visualize")
	pc := session.DB("GoTrafficQueue").C("polyline")
	for ;true; {
		rand1 := randomPosition(currentGeoRegion, *ra)
		rand2 := randomPosition(currentGeoRegion, *ra)
		grr := requestGoogleRouteApi(rand1, rand2)
		//fmt.Println("Position", fmt.Sprintf("%.5f->%.5f, %.5f->%.5f",rand2.Lat, rand1.Lat, rand2.Lng, rand1.Lng))
		if grr.Status != "OK"{
			//fmt.Println("ERR", "STATUS=" + grr.Status)
			switch grr.Status {
			case `ZERO_RESULTS`:
				time.Sleep(time.Second * 3)
				break;
			case `OVER_QUERY_LIMIT`:
				timeRand := 5 + rand.Intn(60 - 5)
				fmt.Println("SLEEP", "OVER_QUERY_LIMIT", "Thread", nodeNum, "Sleep", timeRand, "Second")
				time.Sleep(time.Second * time.Duration(timeRand))
			}
			continue

		}
		//fmt.Println("Route", fmt.Sprintf("Found %d Routes", len(grr.Routes)))
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
						//fmt.Println("New Record", fmt.Sprintf("Add new Data %.5f, %.5f", roadLatLng[j].Lat, roadLatLng[j].Lng))
						newModel := &model.Model{}
						newModel.NewInstance(roadLatLng[j], nil)
						newModel.Append(roadLatLng[j+1])
						err := c.Insert(newModel)
						if err != nil {
							//fmt.Println("Error", err.Error())
						}
					}
				}else{
					//fmt.Println("Found Record", fmt.Sprintf("Found Data %.5f, %.5f", roadLatLng[j].Lat, roadLatLng[j].Lng))
					if !foundModel.ContainParent(roadLatLng[j+1]) && !isSameVector(roadLatLng[j+1], *foundModel){
					//if !foundModel.ContainParent(roadLatLng[j+1]){
						fmt.Println("Found Record", fmt.Sprintf("Add relation %.5f, %.5f to %.5f, %.5f",
							roadLatLng[j+1].Lat, roadLatLng[j+1].Lng,
							roadLatLng[j].Lat, roadLatLng[j].Lng,
						))
						oldModel := foundModel.CopyInstance()
						oldModel.Append(roadLatLng[j+1])
						err := c.Update(foundModel, oldModel)
						if err != nil {
							//fmt.Println("Error", err.Error())
						}
					}else{
						//fmt.Println("Already Relation", "Found Already Relation Road")
					}
				}
			}
		}
		ch <- nodeNum
		time.Sleep(time.Minute * 1)
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