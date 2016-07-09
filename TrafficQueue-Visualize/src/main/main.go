package main

import (
	"github.com/kataras/iris"
	"gopkg.in/mgo.v2"
	"log"
	"gopkg.in/mgo.v2/bson"
)

type Last struct {
	Last string	`bson:last`
}

type LatLng struct {
	Lat	float64	`bson:"lat" json:"lat"`
	Lng	float64	`bson:"lng" json:"lng"`
}

type Model struct {
	Host	LatLng		`bson:"host"`
	Parent	[]LatLng	`bson:"parent"`
}

func main(){
	iris.Get("/map", func (ctx *iris.Context){
		if err := ctx.Render("map.html", nil); err != nil {
			iris.Logger.Printf(err.Error())
		}
	})
	iris.Get("/map/cover", func(ctx *iris.Context){
		if err := ctx.Render("map-cover.html", nil); err != nil {
			iris.Logger.Printf(err.Error())
		}
	})

	iris.StaticWeb("/jquery", "./templates/jquery-3.1.0.min.js", 1)

	iris.Get("/api/intersection", func(ctx *iris.Context){
		session, err := mgo.Dial("127.0.0.1")
		if err != nil {
			panic(err)
		}
		defer session.Close()

		// Optional. Switch the session to a monotonic behavior.
		session.SetMode(mgo.Monotonic, true)

		c := session.DB("GoTrafficQueue").C("road_relate")
		if err != nil {
			log.Fatal(err)
		}

		model := []Model{}
		c.Find(bson.M{"$where":"this.parent.length > 2"}).All(&model)
		ctx.JSON(iris.StatusOK, model)
	})

	iris.Get("/api/last", func(ctx *iris.Context){
		session, err := mgo.Dial("127.0.0.1")
		if err != nil {
			panic(err)
		}
		defer session.Close()

		// Optional. Switch the session to a monotonic behavior.
		session.SetMode(mgo.Monotonic, true)

		c := session.DB("GoTrafficQueue").C("visualize")
		if err != nil {
			log.Fatal(err)
		}

		last := Last{}

		err = c.Find(bson.M{}).One(&last)

		ctx.JSON(iris.StatusOK, last)
	})
	//
	//iris.Get("/", func (ctx *iris.Context){
	//	ctx.Redirect("/map", iris.StatusTemporaryRedirect)
	//})

	iris.Listen(":80")
}