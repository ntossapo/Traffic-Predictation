package mongo

import (
	"gopkg.in/mgo.v2"
	"geo"
	"gopkg.in/mgo.v2/bson"
)

func AddLimitedRegion(session *mgo.Session, ls geo.GeoLimitSquare){
	c := session.DB("GoTrafficQueue").C("region")
	c.RemoveAll(bson.M{"big":1})
	c.Insert(bson.M{"big":1, "geo":ls})
}