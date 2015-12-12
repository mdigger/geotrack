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
	places := []Place{
		Place{Polygon: &geo.Polygon{{
			{37.5667, 55.7152}, {37.5688, 55.7167}, {37.5703, 55.7169},
			{37.5706, 55.7168}, {37.5726, 55.7159}, {37.5728, 55.7158},
			{37.5731, 55.7159}, {37.5751, 55.7152}, {37.5758, 55.7148},
			{37.5755, 55.7144}, {37.5749, 55.7141}, {37.5717, 55.7131},
			{37.5709, 55.7128}, {37.5694, 55.7125}, {37.5661, 55.7145},
			{37.5660, 55.7147}, {37.5667, 55.7152}}}, Name: "Работа"},
		Place{Circle: &geo.Circle{geo.Point{37.589248, 55.765944}, 200.0}, Name: "Дом"},
		Place{Polygon: &geo.Polygon{{
			{37.6256, 55.7522}, {37.6304, 55.7523}, {37.6310, 55.7527},
			{37.6322, 55.7526}, {37.6320, 55.7521}, {37.6326, 55.7517},
			{37.6321, 55.7499}, {37.6305, 55.7499}, {37.6305, 55.7502},
			{37.6264, 55.7504}, {37.6264, 55.7500}, {37.6254, 55.7500},
			{37.6253, 55.7520}, {37.6256, 55.7522}}}, Name: "Знаменский монастырь"},
	}
	for _, place := range places {
		place.GroupID = groupID
		id, err := db.Save(place)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("id:", id.Hex())
	}
	places, err = db.GetAll(groupID)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Places:", len(places))
	// pretty.Println(places)
	track := tracks.TrackData{
		DeviceID: deviceID,
		GroupID:  groupID,
		Time:     time.Now(),
		Location: geo.NewPoint(37.57351, 55.715084),
	}
	placeIDs, err := db.Track(track)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("In %d places\n%v\n", len(placeIDs), placeIDs)
}

func TestPlaces2(t *testing.T) {
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
		groupID = users.SampleGroupID
	)
	places := []Place{
		Place{Polygon: &geo.Polygon{{
			{37.5667, 55.7152}, {37.5688, 55.7167}, {37.5703, 55.7169},
			{37.5706, 55.7168}, {37.5726, 55.7159}, {37.5728, 55.7158},
			{37.5731, 55.7159}, {37.5751, 55.7152}, {37.5758, 55.7148},
			{37.5755, 55.7144}, {37.5749, 55.7141}, {37.5717, 55.7131},
			{37.5709, 55.7128}, {37.5694, 55.7125}, {37.5661, 55.7145},
			{37.5660, 55.7147}, {37.5667, 55.7152}}}, Name: "Работа"},
		Place{Circle: &geo.Circle{geo.Point{37.589248, 55.765944}, 200.0}, Name: "Дом Дмитрия"},
		Place{Circle: &geo.Circle{geo.Point{37.539401, 55.77514}, 200.0}, Name: "Дом Сергея"},
		Place{Circle: &geo.Circle{geo.Point{37.510511, 55.666675}, 200.0}, Name: "Дом Андрея"},
		Place{Polygon: &geo.Polygon{{
			{37.6256, 55.7522}, {37.6304, 55.7523}, {37.6310, 55.7527},
			{37.6322, 55.7526}, {37.6320, 55.7521}, {37.6326, 55.7517},
			{37.6321, 55.7499}, {37.6305, 55.7499}, {37.6305, 55.7502},
			{37.6264, 55.7504}, {37.6264, 55.7500}, {37.6254, 55.7500},
			{37.6253, 55.7520}, {37.6256, 55.7522}}}, Name: "Знаменский монастырь"},
	}
	for _, place := range places {
		place.GroupID = groupID
		id, err := db.Save(place)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("id:", id.Hex())
	}
}
