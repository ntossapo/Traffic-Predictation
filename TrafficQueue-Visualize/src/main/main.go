package main

import (
	"github.com/kataras/iris"
	"gopkg.in/mgo.v2"
	"log"
	"gopkg.in/mgo.v2/bson"
	"unicode/utf8"
	"strconv"
	"geo"
)

type LatLng struct {
	Lat	float64	`bson:"lat" json:"lat"`
	Lng	float64	`bson:"lng" json:"lng"`
}

type Model struct {
	Host	LatLng		`bson:"host"`
	Parent	[]LatLng	`bson:"parent"`
}

type Polyline struct {
	polyline	string	`bson:"polyline" json:"polyline"`
}

func main(){
	//iris.Config.Render.Template.IsDevelopment = true

	iris.Get("/region/node", func (ctx *iris.Context){
		if err := ctx.Render("region-node.html", nil) ; err != nil{
			iris.Logger.Printf(err.Error())
		}
	});

	iris.Get("/region", func (ctx *iris.Context){
		if err := ctx.Render("region.html", nil) ; err != nil{
			iris.Logger.Printf(err.Error())
		}
	});

	iris.Get("/relate/:branch", func (ctx *iris.Context){
		b := ctx.Param("branch")
		if err := ctx.Render("relate.html", map[string]string{"branch":b}) ; err != nil{
			iris.Logger.Printf(err.Error())
		}
	});

	iris.Get("/map/:branch", func (ctx *iris.Context){
		b := ctx.Param("branch")
		if err := ctx.Render("map.html", map[string]string{"branch":b}); err != nil {
			iris.Logger.Printf(err.Error())
		}
	})

	iris.Get("/cover", func(ctx *iris.Context){
		if err := ctx.Render("map-cover.html", nil); err != nil {
			iris.Logger.Printf(err.Error())
		}
	})

	iris.Get("/node/:id", func(ctx *iris.Context){
		id := ctx.Param("id")
		if utf8.RuneCountInString(id) <= 0{
			id = "1"
		}
		if err := ctx.Render("node.html", map[string]string{"node":id}); err != nil {
			iris.Logger.Printf(err.Error())
		}
	})

	iris.Get("/api/region/node/:id", func (ctx *iris.Context){
		id := ctx.Param("id")
		if utf8.RuneCountInString(id) <= 0{
			id = "1"
		}
		intId, _ := strconv.Atoi(id);
		session, err := mgo.Dial("127.0.0.1")
		if err != nil{
			panic(err)
		}
		defer session.Close()

		session.SetMode(mgo.Monotonic, true)

		c := session.DB("GoTrafficQueue").C("worker")
		cc := session.DB("GoTrafficQueue").C("visualize")

		data := struct {
			Node	int
			Limit	geo.GeoLimitSquare
		}{}

		line := struct {
			Node int
			Polyline string
		}{}

		err = c.Find(bson.M{"node":intId}).One(&data)
		if err != nil {
			panic(err)
		}
		err = cc.Find(bson.M{"node":int(intId)}).One(&line)
		if err != nil {
			panic(err)
		}

		ctx.JSON(iris.StatusOK, map[string]interface{}{"node":data.Node, "limit":data.Limit, "line":line.Polyline})

	})

	iris.Get("/api/region/node", func(ctx *iris.Context) {
		session, err := mgo.Dial("127.0.0.1")
		if err != nil{
			panic(err)
		}
		defer session.Close()

		session.SetMode(mgo.Monotonic, true)

		c := session.DB("GoTrafficQueue").C("region")

		data := []struct {
			Big	int
			Geo	geo.GeoLimitSquare
		}{}

		c.Find(bson.M{"big":0}).All(&data)
		ctx.JSON(iris.StatusOK, data)
	})

	iris.Get("/api/region", func(ctx *iris.Context){
		session, err := mgo.Dial("127.0.0.1")
		if err != nil{
			panic(err)
		}
		defer session.Close()

		session.SetMode(mgo.Monotonic, true)

		c := session.DB("GoTrafficQueue").C("region")

		data := struct {
			Big	int
			Geo	geo.GeoLimitSquare
		}{}

		c.Find(bson.M{"big":1}).One(&data)
		ctx.JSON(iris.StatusOK, data)
	})

	iris.Get("/api/coverage", func(ctx *iris.Context){
		session, err := mgo.Dial("127.0.0.1")
		if err != nil {
			panic(err)
		}
		defer session.Close()

		// Optional. Switch the session to a monotonic behavior.
		session.SetMode(mgo.Monotonic, true)

		c := session.DB("GoTrafficQueue").C("polyline")
		if err != nil {
			log.Fatal(err)
		}

		polylines := []map[string]string{}
		c.Find(nil).Select(bson.M{"polyline":1, "_id":0}).All(&polylines)
		ctx.JSON(iris.StatusOK, polylines)
	})

	iris.Get("/api/intersection/:branch", func(ctx *iris.Context){
		session, err := mgo.Dial("127.0.0.1")
		if err != nil {
			panic(err)
		}
		defer session.Close()

		b := ctx.Param("branch")
		if utf8.RuneCountInString(b) <= 0{
			b = "2"
		}
		if _, err := strconv.Atoi(b) ; err != nil{
			b = "2"
		}

		// Optional. Switch the session to a monotonic behavior.
		session.SetMode(mgo.Monotonic, true)

		c := session.DB("GoTrafficQueue").C("road_relate")
		if err != nil {
			log.Fatal(err)
		}

		model := []Model{}
		c.Find(bson.M{"$where":"this.parent.length > " + b}).All(&model)
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

		last := []map[string]string{}

		err = c.Find(bson.M{}).All(&last)

		ctx.JSON(iris.StatusOK, last)
	})
	//
	//iris.Get("/", func (ctx *iris.Context){
	//	ctx.Redirect("/map", iris.StatusTemporaryRedirect)
	//})


	iris.StaticWeb("/jquery", "./templates/jquery-3.1.0.min.js", 1)
	iris.Listen(":80")
}