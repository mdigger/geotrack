package tracks

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/mdigger/geotrack/geo"
	"github.com/mdigger/geotrack/mongo"

	"gopkg.in/mgo.v2/bson"
)

func TestBD(t *testing.T) {
	mdb, err := mongo.Connect("mongodb://localhost/watch")
	if err != nil {
		log.Println("Error connecting to MongoDB:", err)
		return
	}
	defer mdb.Close()

	// подключаемся к хранилищу
	db, err := InitDB(mdb)
	if err != nil {
		t.Fatal("Connect error:", err)
	}

	const deviceID = "test0123456789"
	const count = 156
	for i := 0; i < count; i++ {
		track := &TrackData{
			DeviceID: deviceID,
			Time:     time.Now().Add(time.Minute * time.Duration(-4*(count-i))),
			Point:    geo.NewPoint(55.765944, 37.589248),
		}
		if err := db.Add(track); err != nil {
			t.Fatal(err)
		}
	}

	var lastId bson.ObjectId
	for {
		// fmt.Println("lastID:", lastId.Hex())
		tracks, err := db.Get(deviceID, 5, lastId)
		if err != nil {
			t.Fatal(err)
		}
		if len(tracks) == 0 {
			fmt.Println("END")
			break
		}
		fmt.Println(len(tracks),
			tracks[0].Time.Format("15:04:05"), tracks[0].ID.Hex(), "-",
			tracks[len(tracks)-1].Time.Format("15:04:05"), tracks[len(tracks)-1].ID.Hex())
		lastId = tracks[len(tracks)-1].ID
	}

	track, err := db.GetLast(deviceID)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("last:", track.Time.Format("15:04:05"), track.ID.Hex(), track.Point)
	jsondata, err := json.MarshalIndent(track, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(jsondata))

	deviceIDs, err := db.GetDevicesID()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("device ids:", strings.Join(deviceIDs, ", "))

	_, err = db.GetDay(deviceID)
	if err != nil {
		t.Fatal(err)
	}
}
