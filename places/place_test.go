package places

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
		// {Circle: geo.NewCircle(37.57351, 55.715084, 250.0), Name: "Работа"},
		{Polygon: geo.NewPolygon(
			geo.NewPoint(37.5667, 55.7152),
			geo.NewPoint(37.5688, 55.7167),
			geo.NewPoint(37.5703, 55.7169),
			geo.NewPoint(37.5706, 55.7168),
			geo.NewPoint(37.5726, 55.7159),
			geo.NewPoint(37.5728, 55.7158),
			geo.NewPoint(37.5731, 55.7159),
			geo.NewPoint(37.5751, 55.7152),
			geo.NewPoint(37.5758, 55.7148),
			geo.NewPoint(37.5755, 55.7144),
			geo.NewPoint(37.5749, 55.7141),
			geo.NewPoint(37.5717, 55.7131),
			geo.NewPoint(37.5709, 55.7128),
			geo.NewPoint(37.5694, 55.7125),
			geo.NewPoint(37.5661, 55.7145),
			geo.NewPoint(37.5660, 55.7147),
			geo.NewPoint(37.5667, 55.7152),
		), Name: "Работа"},
		{Circle: geo.NewCircle(37.589248, 55.765944, 200.0), Name: "Дом"},
		{Polygon: geo.NewPolygon(
			geo.NewPoint(37.6256, 55.7522),
			geo.NewPoint(37.6304, 55.7523),
			geo.NewPoint(37.6310, 55.7527),
			geo.NewPoint(37.6322, 55.7526),
			geo.NewPoint(37.6320, 55.7521),
			geo.NewPoint(37.6326, 55.7517),
			geo.NewPoint(37.6321, 55.7499),
			geo.NewPoint(37.6305, 55.7499),
			geo.NewPoint(37.6305, 55.7502),
			geo.NewPoint(37.6264, 55.7504),
			geo.NewPoint(37.6264, 55.7500),
			geo.NewPoint(37.6254, 55.7500),
			geo.NewPoint(37.6253, 55.7520),
			geo.NewPoint(37.6256, 55.7522),
		), Name: "Знаменский монастырь"},
	}
	for _, place := range places {
		place.GroupID = groupID
		id, err := db.Save(place)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("id:", id)
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
