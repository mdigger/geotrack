package place

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/mdigger/geotrack/geo"
	"github.com/mdigger/geotrack/mongo"
	"github.com/mdigger/geotrack/tracks"
	"github.com/mdigger/geotrack/users"
)

func TestPlaces(t *testing.T) {
	mdb, err := mongo.Connect("mongodb://localhost/watch")
	if err != nil {
		log.Println("Error connecting to MongoDB:", err)
		return
	}
	defer mdb.Close()

	db, err := InitDB(mdb)
	if err != nil {
		t.Fatal(err)
	}
	var (
		deviceID = "test0123456789"
		groupID  = users.SampleGroupID
	)
	places := []*Place{
		{Circle: geo.NewCircle(37.57351, 55.715084, 500.0)},
		{Circle: geo.NewCircle(37.589248, 55.765944, 500.0)},
		{Polygon: geo.NewPolygon(
			geo.NewPoint(37.573510, 55.719576),
			geo.NewPoint(37.576561, 55.719234),
			geo.NewPoint(37.579148, 55.718260),
			geo.NewPoint(37.580877, 55.716803),
			geo.NewPoint(37.581484, 55.715084),
			geo.NewPoint(37.580877, 55.713365),
			geo.NewPoint(37.579148, 55.711908),
			geo.NewPoint(37.576561, 55.710934),
			geo.NewPoint(37.573510, 55.710592),
			geo.NewPoint(37.570459, 55.710934),
			geo.NewPoint(37.567872, 55.711908),
			geo.NewPoint(37.566143, 55.713365),
			geo.NewPoint(37.565536, 55.715084),
			geo.NewPoint(37.566143, 55.716803),
			geo.NewPoint(37.567872, 55.718260),
			geo.NewPoint(37.570459, 55.719234),
			geo.NewPoint(37.573510, 55.719576),
		)},
	}
	_ = places
	if err := db.Save(groupID, places...); err != nil {
		t.Fatal(err)
	}
	places, err = db.Get(groupID)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Places:", len(places))
	// pretty.Println(places)
	track := &tracks.TrackData{
		DeviceID: deviceID,
		GroupID:  groupID,
		Time:     time.Now(),
		Point:    geo.NewPoint(37.57351, 55.715084),
	}
	placeIDs, err := db.Track(track)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("In %d places\n%v\n", len(placeIDs), placeIDs)
}
