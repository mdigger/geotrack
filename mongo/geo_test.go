package mongo

import (
	"log"
	"testing"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/kr/pretty"
	"github.com/mdigger/geotrack/geo"
)

var MaxDistance float64 = 4900.0 / geo.EarthRadius // дистанция в радианах

func TestGeo(t *testing.T) {
	mdb, err := Connect("mongodb://localhost/test")
	if err != nil {
		log.Println("Error connecting to MongoDB:", err)
		return
	}
	defer mdb.Close()

	// storeData описывает формат данных для хранения.
	type storeData struct {
		Point geo.Point // координаты
		Time  time.Time // временная метка
	}

	coll := mdb.GetCollection("geo")
	defer mdb.FreeCollection(coll)
	coll.Database.DropDatabase()
	if err = coll.EnsureIndexKey("$2dsphere:point"); err != nil {
		return
	}
	p1 := geo.NewPoint(55.715084, 37.57351)
	p2 := geo.NewPoint(55.765944, 37.589248)
	_ = p1
	_ = p2
	err = coll.Insert(&storeData{
		Point: *p1,
		Time:  time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	var data storeData
	err = coll.Find(bson.M{
		"point": bson.M{
			"$nearSphere":  *p2,
			"$maxDistance": MaxDistance,
		}}).One(&data)
	if err != nil {
		t.Fatal(err)
	}
	pretty.Println(data)
}
